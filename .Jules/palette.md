## 2024-05-18 - Added aria-describedby to setup form
**Learning:** The initial setup form UI had visual helper text (hints) below the inputs, but lacked explicit programmatic linking using `aria-describedby` attributes to associate them for screen readers. Form accessibility must ensure helper text is explicitly connected to the input element it describes.
**Action:** Added `id` attributes to each `.hint` element and added the corresponding `aria-describedby` attribute referencing those IDs on each input. Always link instructional/helper text to forms programmatically.
