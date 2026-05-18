#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Check project-local AI documentation health."""

from __future__ import annotations

import argparse
import fnmatch
import re
import sys
from dataclasses import dataclass
from datetime import date, datetime
from pathlib import Path


ROUTER_COLUMNS = [
    "scope/path_glob",
    "status",
    "must_read",
    "related",
    "keywords",
    "last_verified",
]
VALID_STATUSES = {"active", "stale", "archived", "legacy"}
STALE_DAYS = 30
MEMORY_WARN_BYTES = 12_000
MEMORY_FAIL_BYTES = 20_000
SESSIONS_WARN_BYTES = 40_000
SESSIONS_FAIL_BYTES = 80_000


@dataclass
class CheckResult:
    name: str
    ok: bool
    summary: str
    details: list[str]
    notes: list[str] | None = None


def strip_code(value: str) -> str:
    value = value.strip()
    if len(value) >= 2 and value.startswith("`") and value.endswith("`"):
        return value[1:-1].strip()
    return value


def split_doc_refs(value: str) -> list[str]:
    refs = re.findall(r"`([^`]+)`", value)
    if refs:
        return refs
    return [part for part in re.split(r"\s+", value.strip()) if part and part != "-"]


def repo_root() -> Path:
    return Path(__file__).resolve().parents[1]


def rel_path(path: Path, root: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8-sig")


def parse_router(router_path: Path) -> tuple[list[dict[str, str]], list[str]]:
    errors: list[str] = []
    rows: list[dict[str, str]] = []
    if not router_path.exists():
        return rows, [f"缺少路由表：{router_path}"]

    for line_no, line in enumerate(read_text(router_path).splitlines(), start=1):
        stripped = line.strip()
        if not stripped.startswith("|") or "---" in stripped:
            continue
        cells = [cell.strip() for cell in stripped.strip("|").split("|")]
        if cells == ROUTER_COLUMNS:
            continue
        if len(cells) != len(ROUTER_COLUMNS):
            errors.append(f"第 {line_no} 行列数错误：期望 {len(ROUTER_COLUMNS)} 列，实际 {len(cells)} 列")
            continue
        rows.append(dict(zip(ROUTER_COLUMNS, cells)))
    return rows, errors


def check_router_schema(root: Path, rows: list[dict[str, str]], parse_errors: list[str]) -> CheckResult:
    details = list(parse_errors)
    seen: set[str] = set()
    for idx, row in enumerate(rows, start=1):
        scope = strip_code(row["scope/path_glob"])
        status = row["status"].strip()
        must_read = split_doc_refs(row["must_read"])
        related = split_doc_refs(row["related"])
        last_verified = row["last_verified"].strip()

        if not scope:
            details.append(f"第 {idx} 条 scope/path_glob 为空")
        elif scope in seen:
            details.append(f"重复路由 scope：{scope}")
        seen.add(scope)

        if status not in VALID_STATUSES:
            details.append(f"{scope} 的 status 非法：{status}")

        for doc in must_read + related:
            if doc == "-":
                continue
            if doc.startswith("F:/") or doc.startswith("F:\\"):
                continue
            if not (root / doc).exists():
                details.append(f"{scope} 引用不存在文档：{doc}")

        try:
            datetime.strptime(last_verified, "%Y-%m-%d")
        except ValueError:
            details.append(f"{scope} 的 last_verified 日期非法：{last_verified}")

    if not rows:
        details.append("路由表没有可解析的数据行")

    ok = not details
    summary = f"路由表结构 {'正常' if ok else '异常'}，解析 {len(rows)} 条路由"
    return CheckResult("router_schema", ok, summary, details)


def normalize_context(context: str, root: Path) -> str:
    context_path = Path(context)
    if context_path.is_absolute():
        return rel_path(context_path, root)
    normalized = context_path.as_posix()
    if normalized.startswith("./"):
        return normalized[2:]
    return normalized


def route_matches(scope: str, context: str) -> bool:
    scope = scope.replace("\\", "/").strip("/")
    context = context.replace("\\", "/").strip("/")
    if scope.endswith("/**"):
        prefix = scope[:-3].rstrip("/")
        return context == prefix or context.startswith(prefix + "/")
    if any(ch in scope for ch in "*?[]"):
        return fnmatch.fnmatch(context, scope)
    return context == scope or context.startswith(scope.rstrip("/") + "/")


def check_route_recall(rows: list[dict[str, str]], context: str | None, root: Path) -> CheckResult:
    if not context:
        active_count = sum(1 for row in rows if row["status"].strip() == "active")
        ok = active_count > 0
        details = [] if ok else ["没有 active 路由可用于召回"]
        return CheckResult("route_recall", ok, f"路径路由召回 {'可用' if ok else '不可用'}，active 路由 {active_count} 条", details)

    normalized = normalize_context(context, root)
    matches = [
        row for row in rows
        if row["status"].strip() == "active" and route_matches(strip_code(row["scope/path_glob"]), normalized)
    ]
    details = []
    recall_details = []
    if matches:
        for row in matches:
            scope = strip_code(row["scope/path_glob"])
            must_read = ", ".join(split_doc_refs(row["must_read"]))
            related = ", ".join(split_doc_refs(row["related"]))
            recall_details.append(f"{normalized} -> {scope}；必读：{must_read}；相关：{related}")
    else:
        details.append(f"{normalized} 没有匹配 active 路由")
    ok = bool(matches)
    summary = f"路径路由召回 {'命中' if ok else '未命中'}：{normalized}，匹配 {len(matches)} 条"
    return CheckResult("route_recall", ok, summary, details, recall_details)


def check_stale_docs(root: Path, rows: list[dict[str, str]], today: date) -> CheckResult:
    details: list[str] = []
    for row in rows:
        scope = strip_code(row["scope/path_glob"])
        if row["status"].strip() != "active":
            continue
        try:
            verified = datetime.strptime(row["last_verified"].strip(), "%Y-%m-%d").date()
        except ValueError:
            continue
        age = (today - verified).days
        if age > STALE_DAYS:
            details.append(f"{scope} 已 {age} 天未验证，last_verified={verified.isoformat()}")

    ai_docs = sorted((root / ".ai").glob("*.md"))
    for doc in ai_docs:
        age = (datetime.now() - datetime.fromtimestamp(doc.stat().st_mtime)).days
        if age > STALE_DAYS:
            details.append(f"{rel_path(doc, root)} 文件修改时间已 {age} 天")

    ok = not details
    summary = f"过期文档 {'未发现' if ok else f'发现 {len(details)} 项'}"
    return CheckResult("stale_docs", ok, summary, details)


def check_memory_load(root: Path) -> CheckResult:
    path = root / ".ai" / "MEMORY.md"
    if not path.exists():
        return CheckResult("memory_load", False, "MEMORY 检查失败：文件不存在", [rel_path(path, root)])
    size = path.stat().st_size
    lines = len(read_text(path).splitlines())
    details: list[str] = []
    ok = True
    if size >= MEMORY_FAIL_BYTES:
        ok = False
        details.append(f"MEMORY 已过载：{size} bytes >= {MEMORY_FAIL_BYTES}")
    elif size >= MEMORY_WARN_BYTES:
        details.append(f"MEMORY 接近过载：{size} bytes >= {MEMORY_WARN_BYTES}")
    summary = f"MEMORY 大小 {size} bytes，{lines} 行"
    return CheckResult("memory_load", ok, summary, details)


def check_sessions_length(root: Path) -> CheckResult:
    path = root / ".ai" / "sessions.md"
    if not path.exists():
        return CheckResult("sessions_length", False, "sessions 检查失败：文件不存在", [rel_path(path, root)])
    size = path.stat().st_size
    sessions = sum(1 for line in read_text(path).splitlines() if line.startswith("## "))
    details: list[str] = []
    ok = True
    if size >= SESSIONS_FAIL_BYTES:
        ok = False
        details.append(f"sessions 已过长：{size} bytes >= {SESSIONS_FAIL_BYTES}")
    elif size >= SESSIONS_WARN_BYTES:
        details.append(f"sessions 接近过长：{size} bytes >= {SESSIONS_WARN_BYTES}")
    summary = f"sessions 大小 {size} bytes，归档段落 {sessions} 个"
    return CheckResult("sessions_length", ok, summary, details)


def check_generated_snapshots(root: Path) -> CheckResult:
    path = root / ".ai" / "generated"
    if not path.exists() or not path.is_dir():
        return CheckResult("generated_snapshots", False, "generated 快照目录不存在", [rel_path(path, root)])
    files = [item for item in path.rglob("*") if item.is_file()]
    ok = bool(files)
    details = [] if ok else [f"{rel_path(path, root)} 目录为空，没有生成快照文件"]
    summary = f"generated 快照 {'存在' if ok else '缺失'}，文件数 {len(files)}"
    return CheckResult("generated_snapshots", ok, summary, details)


def print_results(results: list[CheckResult], show_details: bool, only_summary: bool) -> None:
    failures = [result for result in results if not result.ok]
    warnings = [detail for result in results if result.ok for detail in result.details]

    print("AI 文档检查摘要")
    print(f"- 检查项：{len(results)}")
    print(f"- 失败项：{len(failures)}")
    print(f"- 提醒项：{len(warnings)}")
    for result in results:
        mark = "通过" if result.ok else "失败"
        print(f"- [{mark}] {result.summary}")

    if only_summary:
        return

    if show_details:
        print("\n详细信息")
        for result in results:
            lines = result.details + (result.notes or [])
            if not lines:
                continue
            print(f"- {result.name}")
            for detail in lines:
                print(f"  - {detail}")


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="检查 .ai 文档健康状态")
    parser.add_argument("--summary", action="store_true", help="只输出中文摘要")
    parser.add_argument("--context", metavar="PATH", help="检查指定路径的路由召回")
    parser.add_argument("--stale", action="store_true", help="只运行过期文档检查")
    parser.add_argument("--details", action="store_true", help="输出详细问题与召回结果")
    return parser


def main(argv: list[str] | None = None) -> int:
    if hasattr(sys.stdout, "reconfigure"):
        sys.stdout.reconfigure(encoding="utf-8")
    if hasattr(sys.stderr, "reconfigure"):
        sys.stderr.reconfigure(encoding="utf-8")

    args = build_parser().parse_args(argv)
    root = repo_root()
    rows, parse_errors = parse_router(root / ".ai" / "router.md")

    if args.stale:
        results = [check_stale_docs(root, rows, date.today())]
    else:
        results = [
            check_router_schema(root, rows, parse_errors),
            check_route_recall(rows, args.context, root),
            check_stale_docs(root, rows, date.today()),
            check_memory_load(root),
            check_sessions_length(root),
            check_generated_snapshots(root),
        ]

    print_results(results, show_details=args.details, only_summary=args.summary)
    return 0 if all(result.ok for result in results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
