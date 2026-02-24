# PYQ (Previous Year Questions) Data

## Structure
This folder stores the Previous Year Question data used by StudyClaw's Quizmaster mode to weight quiz topics by exam frequency.

```
pyq/
├── README.md              ← (this file)
├── sem3/
│   ├── subject_code.md    ← List of questions per year
│   └── ...
└── sem4/
    └── ...
```

## How to Populate
1. Download your university's past papers (GTU: https://gturesults.in/question-papers/).
2. Send each PDF to StudyClaw in WhatsApp → it auto-indexes it.
3. OR paste questions manually into `sem{N}/{SUBJECT_CODE}.md`.

## Example entry (sem3/3130702.md — Circuit Theory)
```
## 2024
- Q3a: Explain Thevenin's theorem with circuit diagram. (7 marks)
- Q5b: Find current in 4Ω resistor using mesh analysis... (7 marks)

## 2023
- Q2a: State and prove Norton's theorem... (7 marks)
```

## Status
- [ ] Sem 3 PYQs added
- [ ] Sem 4 PYQs added

*(StudyClaw will auto-tag high-frequency topics above 60% recurrence as "HIGH YIELD" for quizzes.)*
