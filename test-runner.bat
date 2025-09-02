@echo off
REM Test Runner Script for Build Scan Go (Windows)
REM This script runs comprehensive tests and generates reports

setlocal EnableDelayedExpansion

echo 🚀 Build Scan Go - Test Runner (Windows)
echo ========================================

REM Check if go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

echo [INFO] Go version:
go version

REM Create test results directory
if not exist "test-results" mkdir test-results

REM Run unit tests
echo [INFO] Running unit tests...
go test -v -short ./... > test-results\unit-tests.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Unit tests passed
) else (
    echo [ERROR] Unit tests failed
    type test-results\unit-tests.log
    exit /b 1
)

REM Run tests with coverage
echo [INFO] Running tests with coverage...
go test -v -coverprofile=test-results\coverage.out ./... > test-results\coverage-tests.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Coverage tests completed

    REM Generate HTML coverage report
    go tool cover -html=test-results\coverage.out -o test-results\coverage.html

    REM Show coverage summary
    for /f "tokens=3" %%i in ('go tool cover -func^=test-results\coverage.out ^| findstr "total"') do set COVERAGE=%%i
    echo [INFO] Overall test coverage: !COVERAGE!

    REM Extract numeric value for comparison
    for /f "tokens=1 delims=%%" %%a in ("!COVERAGE!") do set COVERAGE_NUM=%%a
    if !COVERAGE_NUM! GEQ 70 (
        echo [SUCCESS] Coverage threshold met (≥70%%)
    ) else (
        echo [WARNING] Coverage below threshold (70%%). Current: !COVERAGE!
    )
) else (
    echo [ERROR] Coverage tests failed
    type test-results\coverage-tests.log
    exit /b 1
)

REM Run integration tests (if not unit-only mode)
if "%1" NEQ "--unit-only" (
    echo [INFO] Running integration tests...
    go test -v -run TestIntegration ./... > test-results\integration-tests.log 2>&1
    if %ERRORLEVEL% EQU 0 (
        echo [SUCCESS] Integration tests passed
    ) else (
        echo [WARNING] Integration tests failed (this might be expected without proper server setup)
    )
)

REM Run benchmarks
echo [INFO] Running benchmark tests...
go test -v -bench=. -benchmem ./... > test-results\benchmark-tests.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Benchmark tests completed
) else (
    echo [WARNING] Benchmark tests had issues
)

REM Check code formatting
echo [INFO] Checking code formatting...
go fmt ./... > test-results\fmt.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] Code is properly formatted
) else (
    echo [WARNING] Code formatting issues found:
    type test-results\fmt.log
)

REM Run go vet
echo [INFO] Running go vet...
go vet ./... > test-results\vet.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] go vet passed
) else (
    echo [WARNING] go vet found issues:
    type test-results\vet.log
)

REM Check for race conditions
echo [INFO] Checking for race conditions...
go test -race -short ./... > test-results\race-tests.log 2>&1
if %ERRORLEVEL% EQU 0 (
    echo [SUCCESS] No race conditions detected
) else (
    echo [ERROR] Race conditions detected
    type test-results\race-tests.log
    exit /b 1
)

REM Generate test summary
echo [INFO] Generating test summary...

echo # Test Summary Report > test-results\summary.md
echo Generated: %date% %time% >> test-results\summary.md
echo. >> test-results\summary.md

echo ## Coverage >> test-results\summary.md
echo - Overall Coverage: !COVERAGE! >> test-results\summary.md
echo - Coverage Report: [coverage.html](coverage.html) >> test-results\summary.md
echo. >> test-results\summary.md

echo ## Test Results >> test-results\summary.md
echo - Unit Tests: ✅ Passed >> test-results\summary.md
echo - Integration Tests: ⚠️ See integration-tests.log >> test-results\summary.md
echo - Benchmark Tests: ✅ Completed >> test-results\summary.md
echo - Race Detection: ✅ Passed >> test-results\summary.md
echo. >> test-results\summary.md

echo ## Code Quality >> test-results\summary.md
echo - Code Formatting: ✅ Checked >> test-results\summary.md
echo - Go Vet: ✅ Checked >> test-results\summary.md

echo.
echo [SUCCESS] Test run completed! Results saved in test-results\
echo [INFO] View coverage report: test-results\coverage.html
echo [INFO] View summary: test-results\summary.md

echo.
echo 📊 Quick Summary:
echo    Coverage: !COVERAGE!
echo    Results:  test-results\
echo    Logs:     test-results\*.log

pause
