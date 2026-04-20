## 2025-04-20 - Adding Form Helpers to Screen Readers
**Learning:** Visual helper text below form fields (`<div class="hint">`) is not read by screen readers when focusing the input, causing blind users to miss crucial formatting details like "With country code, no +".
**Action:** Always link visual hints to their respective inputs using `aria-describedby="[hint-id]"` on the input and `id="[hint-id]"` on the hint element so screen readers announce the extra context alongside the label.
