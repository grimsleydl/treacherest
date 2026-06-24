# Template-first devenv standardization

Status: accepted

Treacherest follows a shared devenv convention across projects using a copyable base template before introducing a shared module dependency. The template should standardize command names, process naming, port allocation, readiness checks, and task shape while allowing each repo to adapt language-specific details locally. A shared imported module may be extracted later if several projects converge on the same stable pattern.
