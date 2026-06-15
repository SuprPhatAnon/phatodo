# TOON Reference

This repository uses the TOON specification as the reference format for compact machine-readable output.

- Source: https://raw.githubusercontent.com/toon-format/spec/refs/heads/main/SPEC.md
- Upstream repository: https://github.com/toon-format/spec
- Current spec version: 3.3
- Spec date: 2026-05-21
- Status: Working Draft

## What TOON Is

TOON, short for Token-Oriented Object Notation, is a line-oriented text format for JSON-shaped data.
It is designed to be compact, deterministic, and easy for agents to parse.

## Core Rules

- Documents are UTF-8 text with LF line endings.
- Indentation uses spaces only; tabs must not be used for indentation.
- Objects use indentation instead of braces.
- Arrays declare their length in a header and may include a field list for tabular rows.
- Strings are quoted only when needed.
- Booleans and null use lowercase literals.
- Numbers should be normalized to canonical decimal form when possible.

## Syntax Notes

- Root objects use `key: value` for primitive fields.
- Nested or empty objects use `key:` and continue on indented lines.
- Array headers include the item count, for example `items[2]:`.
- Uniform object arrays may include a field list, for example `items[2]{id,title}:`.
- Delimiters can be comma, tab, or pipe, with the closest array header determining the active delimiter.

## Conformance Notes

- Strict mode validates counts, indentation, delimiter consistency, and escape handling.
- Quoted tokens decode as strings.
- Unquoted tokens may decode as booleans, null, numbers, or strings depending on the token shape.

For implementation details and the full normative text, refer to the upstream spec at the source URL above.
