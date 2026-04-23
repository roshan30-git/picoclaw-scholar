## 2024-05-18 - Link form inputs to hints using `aria-describedby`
**Learning:** Found an accessibility opportunity in `pkg/setup/server.go` where setup form inputs had visible helper text `<div class="hint">` but lacked proper screen reader association.
**Action:** Implemented `aria-describedby` on the input fields and linked them to `id` attributes on the hint divs. Ensure that whenever helper text is provided alongside a form input, `aria-describedby` is utilized to establish the connection for assistive technologies.
