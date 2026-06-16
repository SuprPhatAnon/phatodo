#!/usr/bin/env python3
import os
import subprocess
import sys


MODULE_PREFIX = "github.com/SuprPhatAnon/phatodo"

# Files excluded from per-file coverage thresholds.
# Use this for:
#   - generated sqlc files
#   - command entrypoints that only wire process startup
# Pattern: relative path from repo root, matched exactly.
EXCLUDED_FILES = {
    "cmd/phatodo-server/main.go",
    "cmd/ptodo/main.go",
}

EXCLUDED_DIR_PREFIXES = (
    "internal/storage/postgres/sqlc/",
)

SKIP_DIR_NAMES = {
    ".git",
    ".phatodo",
    ".trekker",
    "bin",
    "node_modules",
    "web",
}

FILE_THRESHOLD_OVERRIDES = {
    # Most Postgres store methods require a live database or broader integration
    # harness to exercise meaningfully. Keep them visible in the report without
    # making this helper unusable while that harness is not present.
    "internal/storage/postgres/audit.go": 0.0,
    "internal/storage/postgres/auth.go": 0.0,
    "internal/storage/postgres/bootstrap.go": 0.0,
    "internal/storage/postgres/comments.go": 0.0,
    "internal/storage/postgres/dependencies.go": 0.0,
    "internal/storage/postgres/epics.go": 0.0,
    "internal/storage/postgres/locks.go": 0.0,
    "internal/storage/postgres/project_config.go": 0.0,
    "internal/storage/postgres/query.go": 0.0,
    "internal/storage/postgres/sqlc_mappers.go": 0.0,
    "internal/storage/postgres/tasks.go": 0.0,
}


def is_critical(file_path):
    return file_path.startswith(("internal/cli/", "internal/server/"))


def find_all_go_files(repo_root):
    go_files = []
    for root, dirs, files in os.walk(repo_root):
        dirs[:] = [d for d in dirs if d not in SKIP_DIR_NAMES]
        rel_root = os.path.relpath(root, repo_root)
        if rel_root == ".":
            rel_root = ""
        if any(rel_root.startswith(prefix.rstrip("/")) for prefix in EXCLUDED_DIR_PREFIXES):
            continue
        for filename in files:
            if not filename.endswith(".go") or filename.endswith("_test.go"):
                continue
            full_path = os.path.join(root, filename)
            rel_path = os.path.relpath(full_path, repo_root)
            go_files.append(rel_path)
    return sorted(go_files)


def is_excluded(file_path, total_statements):
    if file_path in EXCLUDED_FILES:
        return True
    if any(file_path.startswith(prefix) for prefix in EXCLUDED_DIR_PREFIXES):
        return True
    return total_statements == 0


def run_coverage(repo_root):
    cov_file = "/tmp/phatodo_coverage.out"
    if os.path.exists(cov_file):
        os.remove(cov_file)

    env = os.environ.copy()
    env.setdefault("GOCACHE", "/tmp/go-build")

    cmd = [
        "/usr/local/go/bin/go",
        "test",
        "-coverprofile=" + cov_file,
        "./...",
    ]
    result = subprocess.run(cmd, cwd=repo_root, env=env, capture_output=True, text=True)
    if result.stdout:
        print(result.stdout, end="")
    if result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    if result.returncode != 0:
        sys.exit(result.returncode)

    coverage_data = {}
    if not os.path.exists(cov_file):
        return coverage_data

    with open(cov_file, "r", encoding="utf-8") as coverage_file:
        lines = coverage_file.readlines()

    for line in lines[1:]:
        parts = line.strip().split(" ")
        if len(parts) != 3:
            continue

        block, num_stmts_str, count_str = parts
        file_ref = block.split(":")[0]
        if not file_ref.startswith(MODULE_PREFIX):
            continue

        rel_file = file_ref.replace(MODULE_PREFIX + "/", "", 1)
        num_stmts = int(num_stmts_str)
        count = int(count_str)

        if rel_file not in coverage_data:
            coverage_data[rel_file] = {"covered": 0, "total": 0}
        coverage_data[rel_file]["total"] += num_stmts
        if count > 0:
            coverage_data[rel_file]["covered"] += num_stmts

    return coverage_data


def build_report(repo_root, coverage_data):
    rows = []
    for file_path in find_all_go_files(repo_root):
        data = coverage_data.get(file_path, {"covered": 0, "total": 0})
        covered = data["covered"]
        total = data["total"]

        if is_excluded(file_path, total):
            continue

        pct = (covered / total) * 100.0 if total > 0 else 0.0
        critical = is_critical(file_path)
        threshold = FILE_THRESHOLD_OVERRIDES.get(file_path, 80.0 if critical else 70.0)
        status = "PASS" if pct >= threshold else "FAIL"

        rows.append(
            {
                "file": file_path,
                "pct": pct,
                "covered": covered,
                "total": total,
                "is_critical": critical,
                "threshold": threshold,
                "status": status,
            }
        )

    rows.sort(key=lambda row: (not row["is_critical"], row["file"]))
    return rows


def print_summary(rows):
    failed = [row for row in rows if row["status"] == "FAIL"]
    critical_failed = [row for row in failed if row["is_critical"]]
    standard_failed = [row for row in failed if not row["is_critical"]]

    print("\n" + "=" * 80)
    print(" PHATODO TEST COVERAGE SUMMARY")
    print("=" * 80)
    print(f"Total Go Files Analyzed: {len(rows)}")
    print(
        "Failed Coverage Targets: "
        f"{len(failed)} (Critical: {len(critical_failed)}, Standard: {len(standard_failed)})"
    )
    print("=" * 80)

    if failed:
        print("\nFAILED COVERAGE FILES:")
        print(f"{'File Path':<60} | {'Crit?':<5} | {'Coverage':<8} | {'Target':<6}")
        print("-" * 88)
        for row in failed:
            critical = "YES" if row["is_critical"] else "no"
            print(
                f"{row['file']:<60} | {critical:<5} | "
                f"{row['pct']:>7.1f}% | {row['threshold']:>5.1f}%"
            )
    else:
        print("\nAll files pass their coverage thresholds.")

    return failed


def write_markdown_report(repo_root, rows, failed):
    critical_failed = [row for row in failed if row["is_critical"]]
    standard_failed = [row for row in failed if not row["is_critical"]]
    generated_at = subprocess.check_output(["date", "-u"]).decode("utf-8").strip()

    md = [
        "# Phatodo Code Coverage Report\n",
        f"Generated at: {generated_at}\n",
        "## Overview\n",
        "This report lists test coverage for non-test Go source files in Phatodo. ",
        "Generated sqlc files, command entrypoints, and files with no executable statements are excluded. ",
        "Critical CLI/server files use an 80% threshold; other files use 70% unless overridden.\n",
        f"- **Total Go Files**: {len(rows)}",
        (
            f"- **Failed Coverage Targets**: {len(failed)} "
            f"(Critical: {len(critical_failed)}, Standard: {len(standard_failed)})\n"
        ),
        "## Failed Coverage Files\n",
    ]

    if failed:
        md.extend(
            [
                "| File Path | Critical? | Coverage | Target | Status |",
                "|---|---|---|---|---|",
            ]
        )
        for row in failed:
            critical = "**YES**" if row["is_critical"] else "No"
            path = os.path.join(repo_root, row["file"])
            md.append(
                f"| [`{row['file']}`]({path}) | {critical} | "
                f"**{row['pct']:.1f}%** | {row['threshold']:.1f}% | FAIL |"
            )
    else:
        md.append("All files pass their coverage thresholds.\n")

    md.extend(
        [
            "\n## Complete File Coverage Details\n",
            "| File Path | Critical? | Coverage | Target | Status |",
            "|---|---|---|---|---|",
        ]
    )
    for row in rows:
        critical = "**YES**" if row["is_critical"] else "No"
        status = "PASS" if row["status"] == "PASS" else "FAIL"
        pct = f"**{row['pct']:.1f}%**" if row["status"] == "FAIL" else f"{row['pct']:.1f}%"
        path = os.path.join(repo_root, row["file"])
        md.append(
            f"| [`{row['file']}`]({path}) | {critical} | "
            f"{pct} | {row['threshold']:.1f}% | {status} |"
        )

    report_path = os.path.join(repo_root, "coverage_report.md")
    with open(report_path, "w", encoding="utf-8") as report_file:
        report_file.write("\n".join(md))

    print(f"\nFull report written to: {report_path}")


def main():
    repo_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    os.chdir(repo_root)

    print("Running tests and generating coverage profile...")
    coverage_data = run_coverage(repo_root)
    rows = build_report(repo_root, coverage_data)
    failed = print_summary(rows)
    write_markdown_report(repo_root, rows, failed)

    sys.exit(1 if failed else 0)


if __name__ == "__main__":
    main()
