## 2024-04-24 - Form Helper Text Accessibility
**Learning:** Form inputs with visual helper texts (`<div class="hint">`) often lack explicit programmatic association, making it difficult for screen reader users to understand expected input formats or where to get tokens.
**Action:** Always add explicit `aria-describedby` attributes to `<input>` elements that point to the `id` of their corresponding visual helper text elements.
