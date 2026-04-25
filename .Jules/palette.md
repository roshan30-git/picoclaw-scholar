## 2024-04-19 - Accessibility for Form Hints
**Learning:** The setup form had `.hint` divs below inputs, but screen readers wouldn't associate the hint text with the input fields. This is a common accessibility issue for forms with help text.
**Action:** Use `aria-describedby="hint-id"` on the input element and add `id="hint-id"` to the corresponding hint element to programmatically link them for assistive technologies. Adding an `aria-hidden="true"` visual required indicator `*` also helps users quickly identify mandatory fields without confusing screen readers (which already announce the `required` attribute).

## 2024-05-18 - Keyboard Navigation in Top Navbar
**Learning:** The "View Source" button on the docs page had `focus:outline-none`, completely removing keyboard focus visibility. This makes it impossible for screen reader or keyboard-only users to know when they are focused on this important primary action link.
**Action:** Replaced it with `focus:outline-none focus-visible:ring-2 focus-visible:ring-primary`, ensuring focus outlines appear for keyboard nav while maintaining the clean un-outlined look for mouse clicks.
