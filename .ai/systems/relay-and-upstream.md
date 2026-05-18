# Relay And Upstream

The active gateway logic lives under:

- `backend/internal/handler/`
- `backend/internal/service/`
- `backend/internal/pkg/openai/`
- `backend/internal/pkg/openai_compat/`

Important production validation paths:

- Authenticated `GET /v1/models`
- `POST /v1/chat/completions`
- `/v1/messages` if Anthropic-compatible clients are involved
- `/v1beta/*` if Gemini-compatible clients are involved

When debugging routing:

1. Identify API key group.
2. Verify group platform.
3. Verify account bindings and active schedulable accounts.
4. Verify Redis concurrency writes.
5. Check app logs for `account_select_failed`, `forward_failed`, or upstream status codes.

