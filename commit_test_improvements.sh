#!/bin/bash

# Script to commit all the framework test improvements

echo "Staging changes..."

# Stage all modified files
git add gra.go
git add CHANGELOG.md
git add TEST_COVERAGE.md
git add TEST_PLAN.md
git add BENCHMARK.md
git add Makefile
git add bench.out
git add coverage.out
git add coverage.html
git add logger/logger.go
git add logger/logger_test.go
git add validator/validator.go
git add validator/validator_test.go
git add router/router_bench_test.go
git add router/router_test.go
git add scripts/run_tests_with_coverage.sh

# Commit with a descriptive message
git commit -m "Enhance testing and improve coverage to 91.7%

- Fixed fatal logger tests by adding mockable osExit
- Fixed array/slice validation in validator tests
- Added benchmarking results and documentation
- Created BENCHMARK.md with performance analysis
- Updated TEST_COVERAGE.md with improved metrics
- Ran benchmarks and included results in repo
- Overall coverage increased from 90.2% to 91.7%"

echo "Changes committed successfully!"
echo "You can now push these changes to GitHub with: git push origin main"
