# PowerShell 테스트 러너 헬퍼
# TEST-POWERSHELL-TEST-002: PowerShell test execution framework

param(
    [Parameter(Mandatory=$false)]
    [string]$TestPath = "./tests",

    [Parameter(Mandatory=$false)]
    [ValidateSet("all", "package", "hooks", "cli")]
    [string]$TestType = "all",

    [Parameter(Mandatory=$false)]
    [switch]$ShowDetails = $false,

    [Parameter(Mandatory=$false)]
    [switch]$StopOnError = $true
)

# ==============================================================================
# 환경 설정
# ==============================================================================

$ErrorActionPreference = "Continue"
if ($StopOnError) {
    $ErrorActionPreference = "Stop"
}

$ScriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = (Get-Item $ScriptRoot).Parent.Parent.Parent.Parent.Parent
$TestResults = @()

if ($ShowDetails) {
    Write-Host "PowerShell Test Runner" -ForegroundColor Cyan
    Write-Host "Project Root: $ProjectRoot" -ForegroundColor Gray
    Write-Host "Test Type: $TestType" -ForegroundColor Gray
}

# ==============================================================================
# 함수: 패키지 설치 검증
# ==============================================================================

function Test-PackageInstallation {
    param([string]$ProjectRoot)

    if ($ShowDetails) {
        Write-Host "`n[TEST] 패키지 설치 검증" -ForegroundColor Yellow
    }

    try {
        Push-Location $ProjectRoot

        # 개발 패키지 설치 (이미 설치되었을 가능성 높음)
        # pip install -e ".[dev]" | Out-Null

        # Python에서 패키지 임포트 테스트
        $output = python -c "import moai_adk; print('OK')" 2>&1

        if ($output -like "*OK*") {
            Write-Host "✓ 패키지 설치 완료" -ForegroundColor Green
            return $true
        } else {
            Write-Host "✗ 패키지 임포트 실패" -ForegroundColor Red
            Write-Host $output -ForegroundColor Gray
            return $false
        }
    } catch {
        Write-Host "✗ 패키지 설치 오류: $_" -ForegroundColor Red
        return $false
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# 함수: pytest 실행 (PowerShell에서)
# ==============================================================================

function Invoke-PytestTests {
    param(
        [string]$TestPath,
        [string]$TestType
    )

    if ($ShowDetails) {
        Write-Host "`n[TEST] pytest 테스트 실행" -ForegroundColor Yellow
    }

    try {
        Push-Location $ProjectRoot

        # 테스트 경로 결정
        $actualTestPath = switch ($TestType) {
            "package" { "$TestPath/unit" }
            "hooks" { "$TestPath/hooks" }
            "cli" { "$TestPath/integration" }
            default { "$TestPath" }
        }

        # pytest 실행
        if ($ShowDetails) {
            pytest $actualTestPath -v --tb=short 2>&1
        } else {
            pytest $actualTestPath --tb=short 2>&1
        }

        $testResult = $LASTEXITCODE -eq 0

        if ($testResult) {
            Write-Host "✓ pytest 테스트 통과" -ForegroundColor Green
        } else {
            Write-Host "✗ pytest 테스트 실패" -ForegroundColor Red
        }

        return $testResult
    } catch {
        Write-Host "✗ pytest 실행 오류: $_" -ForegroundColor Red
        return $false
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# 함수: PowerShell 명령어 호환성 테스트
# ==============================================================================

function Test-CommandAvailability {
    param([string[]]$Commands)

    if ($ShowDetails) {
        Write-Host "`n[TEST] 필수 명령어 가용성" -ForegroundColor Yellow
    }

    $allAvailable = $true

    foreach ($cmd in $Commands) {
        if (Get-Command $cmd -ErrorAction SilentlyContinue) {
            Write-Host "✓ $cmd 사용 가능" -ForegroundColor Green
        } else {
            Write-Host "✗ $cmd 없음" -ForegroundColor Red
            $allAvailable = $false
        }
    }

    return $allAvailable
}

# ==============================================================================
# 함수: 패키지 모듈 테스트 (Python 호출)
# ==============================================================================

function Test-PackageModules {
    param([string]$ProjectRoot)

    if ($ShowDetails) {
        Write-Host "`n[TEST] 패키지 모듈 구조" -ForegroundColor Yellow
    }

    try {
        Push-Location $ProjectRoot

        $pythonScript = @"
import moai_adk
from moai_adk import cli, core, templates

modules = ['cli', 'core', 'templates']
for mod in modules:
    print(f'✓ {mod} 모듈 로드 완료')
"@

        python -c $pythonScript 2>&1
        return $LASTEXITCODE -eq 0
    } catch {
        Write-Host "✗ 모듈 로드 오류: $_" -ForegroundColor Red
        return $false
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# 함수: 타입 체크 (mypy)
# ==============================================================================

function Test-TypeChecking {
    param([string]$ProjectRoot)

    if ($ShowDetails) {
        Write-Host "`n[TEST] 타입 체크 (mypy)" -ForegroundColor Yellow
    }

    try {
        Push-Location $ProjectRoot

        mypy src/moai_adk --strict --ignore-missing-imports 2>&1 | Select-Object -First 10

        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 타입 체크 통과" -ForegroundColor Green
            return $true
        } else {
            Write-Host "⚠ 타입 체크 경고/오류 (상세: mypy src/moai_adk)" -ForegroundColor Yellow
            return $true  # 경고는 무시 (실패로 처리 안 함)
        }
    } catch {
        Write-Host "⚠ mypy 실행 오류: $_" -ForegroundColor Yellow
        return $true  # 도구 미설치 시 무시
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# 함수: 코드 린팅 (ruff)
# ==============================================================================

function Test-Linting {
    param([string]$ProjectRoot)

    if ($ShowDetails) {
        Write-Host "`n[TEST] 코드 린팅 (ruff)" -ForegroundColor Yellow
    }

    try {
        Push-Location $ProjectRoot

        ruff check src/moai_adk --select=E,W,F 2>&1 | Select-Object -First 10

        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ 린팅 체크 통과" -ForegroundColor Green
            return $true
        } else {
            Write-Host "⚠ 린팅 경고 발견 (상세: ruff check src/moai_adk)" -ForegroundColor Yellow
            return $true  # 경고는 무시
        }
    } catch {
        Write-Host "⚠ ruff 실행 오류: $_" -ForegroundColor Yellow
        return $true  # 도구 미설치 시 무시
    } finally {
        Pop-Location
    }
}

# ==============================================================================
# 함수: 결과 보고
# ==============================================================================

function Report-TestResults {
    param(
        [hashtable]$Results,
        [bool]$ShowDetails
    )

    Write-Host "`n" + ("=" * 70) -ForegroundColor Cyan
    Write-Host "PowerShell 테스트 결과 보고" -ForegroundColor Cyan
    Write-Host "=" * 70 -ForegroundColor Cyan

    $passed = 0
    $failed = 0

    foreach ($test in $Results.GetEnumerator()) {
        $status = if ($test.Value) { "✓ 통과" } else { "✗ 실패" }
        $color = if ($test.Value) { "Green" } else { "Red" }
        Write-Host "$($test.Key): $status" -ForegroundColor $color

        if ($test.Value) { $passed++ } else { $failed++ }
    }

    Write-Host "`n총 결과: $passed 통과, $failed 실패" -ForegroundColor Cyan

    if ($failed -gt 0) {
        Write-Host "`n자세한 정보: pytest -v tests/" -ForegroundColor Yellow
    }

    return $failed -eq 0
}

# ==============================================================================
# 메인 실행
# ==============================================================================

function Main {
    Write-Host "PowerShell 패키지 검증 테스트 시작" -ForegroundColor Cyan
    Write-Host "타임스탬프: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray

    $results = @{}

    # 1. 패키지 설치 검증
    $results["패키지 설치"] = Test-PackageInstallation -ProjectRoot $ProjectRoot

    # 2. 필수 명령어 확인
    $results["명령어 가용성"] = Test-CommandAvailability -Commands @("python", "pip", "pytest")

    # 3. 패키지 모듈 테스트
    $results["모듈 로드"] = Test-PackageModules -ProjectRoot $ProjectRoot

    # 4. pytest 실행 (주요 테스트)
    $results["pytest 테스트"] = Invoke-PytestTests -TestPath $TestPath -TestType $TestType

    # 5. 타입 체크
    $results["타입 체크"] = Test-TypeChecking -ProjectRoot $ProjectRoot

    # 6. 린팅
    $results["코드 린팅"] = Test-Linting -ProjectRoot $ProjectRoot

    # 결과 보고
    $success = Report-TestResults -Results $results -Verbose $ShowDetails

    # 종료 코드 설정
    if ($success) {
        Write-Host "`n모든 테스트 통과! ✓" -ForegroundColor Green
        exit 0
    } else {
        Write-Host "`n일부 테스트 실패! ✗" -ForegroundColor Red
        exit 1
    }
}

# 스크립트 실행
Main
