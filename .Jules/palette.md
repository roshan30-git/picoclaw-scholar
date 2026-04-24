## 2025-05-18 - Form Input Hint Accessibility
**Learning:** Using `aria-describedby` to explicitly link helper/hint texts to form `<input>` fields ensures that screen readers announce this supplementary information to users when the input receives focus. Visual proximity alone is insufficient for accessibility.
**Action:** When creating or modifying forms with helper text or hints, always assign a unique `id` to the helper text element and use `aria-describedby="[helper-id]"` on the corresponding `<input>`.
