## 2024-04-19 - Accessibility for Form Hints
**Learning:** The setup form had `.hint` divs below inputs, but screen readers wouldn't associate the hint text with the input fields. This is a common accessibility issue for forms with help text.
**Action:** Use `aria-describedby="hint-id"` on the input element and add `id="hint-id"` to the corresponding hint element to programmatically link them for assistive technologies. Adding an `aria-hidden="true"` visual required indicator `*` also helps users quickly identify mandatory fields without confusing screen readers (which already announce the `required` attribute).
## 2024-05-10 - Accessibility for Interactive Icons
**Learning:** Icon-only elements (like copy buttons) that are implemented as `<span>` tags with `cursor-pointer` lack keyboard accessibility and screen reader support. This makes them unusable for keyboard-only users and unclear to screen reader users.
**Action:** Wrap interactive icons in `<button>` tags, include an `aria-label` or `title`, and add `focus-visible` styles (e.g., `focus-visible:ring-2`) to ensure they are fully navigable and clear to assistive technologies. Added an active state copy feedback to improve user experience.
## 2024-05-15 - Accessibility for Tailwind Interactive Elements
**Learning:** Modern CSS resets and utilities like Tailwind often strip default browser focus outlines. Interactive elements (like `<a>` tags acting as buttons or navigation links) can lose clear visual focus, making the site difficult to navigate for keyboard users.
**Action:** Always explicitly apply `focus-visible` styling (e.g., `focus:outline-none focus-visible:ring-2 focus-visible:ring-primary`) to interactive elements, such as links and CTAs, to ensure clear visual feedback for keyboard navigation without impacting mouse users.
## 2025-05-18 - Form Input Hint Accessibility
**Learning:** Using `aria-describedby` to explicitly link helper/hint texts to form `<input>` fields ensures that screen readers announce this supplementary information to users when the input receives focus. Visual proximity alone is insufficient for accessibility.
**Action:** When creating or modifying forms with helper text or hints, always assign a unique `id` to the helper text element and use `aria-describedby="[helper-id]"` on the corresponding `<input>`. Use systematic IDs like `hint-FIELD_NAME` to keep associations clear and maintainable.
## 2025-05-19 - Keyboard Accessibility for CSS Links
**Learning:** When using custom CSS templates (like in blog articles) or utility frameworks, `<a>` tags might lack distinct focus rings for keyboard navigation.
**Action:** Explicitly define `a:focus-visible` styles with a clear outline or box-shadow (e.g., `box-shadow: 0 0 0 2px var(--bg), 0 0 0 4px var(--teal);`) to ensure keyboard accessibility without affecting mouse hover states.
## 2025-06-29 - Accessibility for Scrollable Regions
**Learning:** Horizontally scrollable containers (using `overflow-x-auto` or `overflow-auto`, common for tables and code blocks) are inaccessible to keyboard users unless they are explicitly focusable. A screen reader also needs context on what the region contains.
**Action:** Always add `tabindex="0"`, a `role="region"`, and an appropriate `aria-label` to containers that use overflow for scrolling. Ensure they also have clear focus styling (e.g., `focus:outline-none focus-visible:ring-2` for Tailwind or custom `:focus-visible` CSS) so the focus indicator is visible without breaking mouse interaction.
