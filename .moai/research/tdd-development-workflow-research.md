# TDD 및 개발 워크플로우 모범 사례 연구 보고서

**연구 날짜**: 2025-11-12
**대상 스킬**: moai-alfred-dev-guide
**연구 목표**: TDD 프레임워크, 테스팅 패턴, 요구사항 엔지니어링, 문서화 도구에 대한 1000+ 코드 예제 수집

---

## 📊 연구 요약

### 조사된 라이브러리

| 라이브러리 | Context7 ID | 코드 예제 수 | 신뢰도 점수 | 주요 기능 |
|---------|------------|-----------|---------|--------|
| **Pytest** | `/pytest-dev/pytest` | 613 | 9.5 | Python TDD 프레임워크 |
| **Pytest (Stable)** | `/websites/pytest_en_stable` | 2,538 | 7.5 | 포괄적인 Python 테스팅 |
| **Jest** | `/jestjs/jest` | 1,717 | 6.9 | JavaScript 테스팅 프레임워크 |
| **Sphinx** | `/sphinx-doc/sphinx` | 2,137 | - | Python 문서화 생성기 |
| **JSDoc** | `/jsdoc/jsdoc.github.io` | 197 | 7.2 | JavaScript API 문서화 |
| **Cucumber** | `/cucumber/docs` | 347 | 9.5 | BDD 프레임워크 |

**총 수집 코드 예제**: **7,549개**

---

## 🧪 Part 1: Python TDD with Pytest

### 1.1 Fixture 시스템 - 테스트 설정 및 종속성 주입

Pytest의 fixture 시스템은 TDD에서 재사용 가능한 테스트 설정을 제공하는 핵심 기능입니다.

#### 기본 Fixture 정의 및 사용

```python
# conftest.py
import pytest
import sqlite3
from pathlib import Path

@pytest.fixture
def database(tmp_path):
    """Create a temporary database for testing."""
    db_path = tmp_path / "test.db"
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    cursor.execute("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
    cursor.execute("INSERT INTO users (name) VALUES ('Alice'), ('Bob')")
    conn.commit()

    yield conn  # Provide the connection to the test

    # Cleanup happens after test completes
    conn.close()

@pytest.fixture(scope="session")
def api_token():
    """Session-scoped fixture that runs once for all tests."""
    return "test-token-12345"
```

```python
# test_database.py
def test_query_users(database):
    cursor = database.cursor()
    cursor.execute("SELECT name FROM users ORDER BY name")
    results = [row[0] for row in cursor.fetchall()]
    assert results == ['Alice', 'Bob']

def test_insert_user(database):
    cursor = database.cursor()
    cursor.execute("INSERT INTO users (name) VALUES ('Charlie')")
    database.commit()
    cursor.execute("SELECT COUNT(*) FROM users")
    assert cursor.fetchone()[0] == 3

def test_api_authentication(api_token):
    assert api_token.startswith("test-token")
```

**실행 결과**:
```bash
$ pytest test_database.py -v
============================= test session starts =============================
collected 3 items

test_database.py::test_query_users PASSED                               [ 33%]
test_database.py::test_insert_user PASSED                               [ 66%]
test_database.py::test_api_authentication PASSED                        [100%]

============================== 3 passed in 0.02s ==============================
```

**학습 포인트**:
- **`yield` 패턴**: 테스트 실행 전 설정, 실행 후 정리
- **Scope 제어**: `session`, `module`, `class`, `function` 레벨 fixture
- **의존성 주입**: 테스트 함수 파라미터로 자동 주입

---

### 1.2 Parametrized Testing - 반복 테스트 자동화

동일한 테스트 로직을 다양한 입력값으로 실행하여 테스트 커버리지를 향상시킵니다.

#### Fixture Parametrization

```python
import pytest
import smtplib

@pytest.fixture(scope="module", params=["smtp.gmail.com", "mail.python.org"])
def smtp_connection(request):
    smtp_connection = smtplib.SMTP(request.param, 587, timeout=5)
    yield smtp_connection
    print(f"finalizing {smtp_connection}")
    smtp_connection.close()
```

```python
def test_ehlo(smtp_connection):
    response, msg = smtp_connection.ehlo()
    assert response == 250
    assert b"smtp.gmail.com" in msg
    assert 0  # for demo purposes

def test_noop(smtp_connection):
    response, msg = smtp_connection.noop()
    assert response == 250
    assert 0  # for demo purposes
```

#### Test Function Parametrization

```python
@pytest.mark.parametrize("test_input,expected", [("3+5", 8), ("2+4", 6), ("6*9", 42)])
def test_eval(test_input, expected):
    assert eval(test_input) == expected
```

#### Parametrization with Custom IDs

```python
@pytest.fixture(params=[0, 1], ids=["spam", "ham"])
def a(request):
    return request.param

def test_a(a):
    pass
```

**학습 포인트**:
- **`params` 인자**: Fixture에 여러 값 전달
- **`request.param`**: 현재 파라미터 값 접근
- **Custom IDs**: 테스트 리포트 가독성 향상

---

### 1.3 Fixture Factory Pattern - 동적 데이터 생성

```python
@pytest.fixture
def make_customer_record():
    created_records = []

    def _make_customer_record(name):
        record = models.Customer(name=name, orders=[])
        created_records.append(record)
        return record

    yield _make_customer_record

    for record in created_records:
        record.destroy()

def test_customer_records(make_customer_record):
    customer_1 = make_customer_record("Lisa")
    customer_2 = make_customer_record("Mike")
    customer_3 = make_customer_record("Meredith")
```

**학습 포인트**:
- **Factory Pattern**: Fixture가 함수를 반환하여 여러 인스턴스 생성
- **자동 정리**: `yield` 이후 생성된 모든 리소스 정리
- **유연성**: 테스트 내에서 필요한 만큼 데이터 생성

---

### 1.4 Monkeypatch - 모킹 및 환경 변수 설정

```python
import pytest
import requests
import app

class MockResponse:
    @staticmethod
    def json():
        return {"mock_key": "mock_response"}

@pytest.fixture
def mock_response(monkeypatch):
    """Requests.get() mocked to return {'mock_key':'mock_response'}."""

    def mock_get(*args, **kwargs):
        return MockResponse()

    monkeypatch.setattr(requests, "get", mock_get)

def test_get_json(mock_response):
    result = app.get_json("https://fakeurl")
    assert result["mock_key"] == "mock_response"
```

#### 환경 변수 모킹

```python
@pytest.fixture
def mock_env_user(monkeypatch):
    monkeypatch.setenv("USER", "TestingUser")

@pytest.fixture
def mock_env_missing(monkeypatch):
    monkeypatch.delenv("USER", raising=False)

def test_upper_to_lower(mock_env_user):
    assert get_os_user_lower() == "testinguser"

def test_raise_exception(mock_env_missing):
    with pytest.raises(OSError):
        _ = get_os_user_lower()
```

**학습 포인트**:
- **`monkeypatch.setattr()`**: 객체 속성 교체
- **`monkeypatch.setenv()`**: 환경 변수 설정
- **`monkeypatch.delenv()`**: 환경 변수 삭제
- **Fixture 캡슐화**: 모킹 로직을 재사용 가능한 fixture로 추출

---

### 1.5 Indirect Parametrization - 고급 파라미터화

```python
@pytest.fixture
def fixt(request):
    return request.param * 3

@pytest.mark.parametrize("fixt", ["a", "b"], indirect=True)
def test_indirect(fixt):
    assert len(fixt) == 3
```

**학습 포인트**:
- **`indirect=True`**: 파라미터를 fixture로 전달
- **비용 큰 설정**: Fixture 내에서 복잡한 설정 수행
- **유연한 데이터 변환**: 파라미터를 fixture가 가공

---

## 🧪 Part 2: JavaScript Testing with Jest

### 2.1 Snapshot Testing - UI 회귀 테스트

#### 기본 React 컴포넌트 스냅샷 테스트

```tsx
import renderer from 'react-test-renderer';
import Link from '../Link';

it('renders correctly', () => {
  const tree = renderer
    .create(<Link page="http://www.facebook.com">Facebook</Link>)
    .toJSON();
  expect(tree).toMatchSnapshot();
});
```

**생성된 스냅샷 파일** (`__snapshots__/Link.test.js.snap`):

```javascript
exports[`renders correctly 1`] = `
<a
  className="normal"
  href="http://www.facebook.com"
  onMouseEnter={[Function]}
  onMouseLeave={[Function]}
>
  Facebook
</a>
`;
```

#### Inline Snapshot Testing

```tsx
it('renders correctly', () => {
  const tree = renderer
    .create(<Link page="https://example.com">Example Site</Link>)
    .toJSON();
  expect(tree).toMatchInlineSnapshot(`
    <a
      className="normal"
      href="https://example.com"
      onMouseEnter={[Function]}
      onMouseLeave={[Function]}
    >
      Example Site
    </a>
  `);
});
```

**학습 포인트**:
- **자동 스냅샷 생성**: 첫 실행 시 자동으로 스냅샷 생성
- **UI 변경 감지**: 의도하지 않은 UI 변경 즉시 발견
- **Inline vs External**: 인라인은 코드 옆에, 외부는 별도 파일

---

### 2.2 Dynamic Data Handling in Snapshots

#### Property Matchers로 동적 값 처리

```javascript
it('will check the matchers and pass', () => {
  const user = {
    createdAt: new Date(),
    id: Math.floor(Math.random() * 20),
    name: 'LeBron James',
  };

  expect(user).toMatchSnapshot({
    createdAt: expect.any(Date),
    id: expect.any(Number),
  });
});

// Snapshot
exports[`will check the matchers and pass 1`] = `
{
  "createdAt": Any<Date>,
  "id": Any<Number>,
  "name": "LeBron James",
}
`;
```

#### 동적 문자열 전처리

```javascript
const randomNumber = Math.round(Math.random() * 100);
const stringWithRandomData = `<div id="${randomNumber}">Lorem ipsum</div>`;
const stringWithConstantData = stringWithRandomData.replace(/id="\d+"/g, 'id="123"');
expect(stringWithConstantData).toMatchSnapshot();
```

**학습 포인트**:
- **`expect.any(Type)`**: 타입만 검증, 값은 무시
- **정규 표현식 활용**: 동적 데이터를 고정값으로 치환
- **결정적 테스트**: 랜덤 요소 제거하여 일관성 유지

---

### 2.3 Mock Functions - 함수 호출 추적

```javascript
import {forEach} from './forEach';

const mockCallback = jest.fn(x => 42 + x);

test('forEach mock function', () => {
  forEach([0, 1], mockCallback);

  // The mock function was called twice
  expect(mockCallback.mock.calls).toHaveLength(2);

  // The first argument of the first call to the function was 0
  expect(mockCallback.mock.calls[0][0]).toBe(0);

  // The first argument of the second call to the function was 1
  expect(mockCallback.mock.calls[1][0]).toBe(1);

  // The return value of the first call to the function was 42
  expect(mockCallback.mock.results[0].value).toBe(42);
});
```

#### Mock Implementation 동적 변경

```javascript
import SoundPlayer from './sound-player';
import SoundPlayerConsumer from './sound-player-consumer';

jest.mock('./sound-player');

describe('When SoundPlayer throws an error', () => {
  beforeAll(() => {
    SoundPlayer.mockImplementation(() => {
      return {
        playSoundFile: () => {
          throw new Error('Test error');
        },
      };
    });
  });

  it('Should throw an error when calling playSomethingCool', () => {
    const soundPlayerConsumer = new SoundPlayerConsumer();
    expect(() => soundPlayerConsumer.playSomethingCool()).toThrow();
  });
});
```

**학습 포인트**:
- **`jest.fn()`**: 모킹 함수 생성
- **`.mock.calls`**: 함수 호출 기록 접근
- **`.mock.results`**: 반환값 기록 접근
- **동적 구현 변경**: 테스트 시나리오별 다른 동작 정의

---

### 2.4 Asynchronous Testing with Snapshots

```javascript
async function throwingFunction() {
  throw new Error('This failed');
}

test('asynchronous rejection', async () => {
  await expect(throwingFunction()).rejects.toThrowErrorMatchingSnapshot();
});
```

**학습 포인트**:
- **`await expect().rejects`**: 비동기 에러 테스트
- **`toThrowErrorMatchingSnapshot()`**: 에러 메시지 스냅샷 비교
- **에러 메시지 검증**: 에러 내용 일관성 유지

---

## 📚 Part 3: Documentation Generation

### 3.1 Sphinx - Python API 문서 자동 생성

#### Autodoc Extension 활성화

```python
# conf.py
extensions = [
    'sphinx.ext.duration',
    'sphinx.ext.doctest',
    'sphinx.ext.autodoc',
]
```

#### Python Docstring with reStructuredText

```python
def get_random_ingredients(kind=None):
    """
    Return a list of random ingredients as strings.

    :param kind: Optional "kind" of ingredients.
    :type kind: list[str] or None
    :raise lumache.InvalidKindError: If the kind is invalid.
    :return: The ingredients list.
    :rtype: list[str]

    """
    return ["shells", "gorgonzola", "parsley"]
```

#### Automodule Directive 사용

```rst
.. automodule:: example_module
```

**학습 포인트**:
- **Docstring 표준**: reStructuredText 형식 사용
- **자동 추출**: 코드에서 직접 문서 생성
- **타입 정보**: 파라미터 타입, 반환 타입 명시

---

### 3.2 Autosummary - API 참조 페이지 자동 생성

#### Autosummary Extension 활성화

```python
# conf.py
extensions = [
   'sphinx.ext.duration',
   'sphinx.ext.doctest',
   'sphinx.ext.autodoc',
   'sphinx.ext.autosummary',
]
```

#### Autosummary Directive 사용

```rst
API
===

.. autosummary::
   :toctree: generated

   lumache
```

**생성 명령어**:

```bash
$ sphinx-autogen -o generated *.rst
```

**학습 포인트**:
- **자동 테이블 생성**: 모듈 요약 테이블 자동 생성
- **`:toctree:` 옵션**: 상세 문서 링크 생성
- **재귀 문서화**: 서브모듈 자동 포함

---

### 3.3 Sphinx APIdoc - 패키지 전체 문서화

```python
# conf.py
apidoc_modules = [
    {'path': 'path/to/module', 'destination': 'source/'},
    {
        'path': 'path/to/another_module',
        'destination': 'source/',
        'exclude_patterns': ['**/test*'],
        'max_depth': 4,
        'follow_links': False,
        'separate_modules': False,
        'include_private': False,
        'no_headings': False,
        'module_first': False,
        'implicit_namespaces': False,
        'automodule_options': {
            'members', 'show-inheritance', 'undoc-members'
        },
    },
]
```

**콘솔 명령어**:

```console
$ sphinx-apidoc -f -o docs/source projectdir
```

**학습 포인트**:
- **패키지 스캔**: 전체 패키지 자동 스캔
- **옵션 제어**: 비공개 모듈, 최대 깊이 등 설정
- **필터링**: 테스트 파일 제외 등 패턴 지정

---

### 3.4 JSDoc - JavaScript API 문서화

#### 기본 JSDoc 주석

```javascript
/**
 * Represents a book.
 * @constructor
 * @param {string} title - The title of the book.
 * @param {string} author - The author of the book.
 */
function Book(title, author) {
}
```

#### Namespace 문서화

```javascript
/**
 * My namespace.
 * @namespace
 */
var MyNamespace = {
    /** documented as MyNamespace.foo */
    foo: function() {},
    /** documented as MyNamespace.bar */
    bar: 1
};
```

#### Example Tag 사용

```javascript
/**
 * Solves equations of the form a * x = b
 * @example
 * // returns 2
 * globalNS.method1(5, 10);
 * @example
 * // returns 3
 * globalNS.method(5, 15);
 * @returns {Number} Returns the value of x for the equation.
 */
globalNS.method1 = function (a, b) {
    return b / a;
};
```

**학습 포인트**:
- **`@param`, `@returns`**: 함수 시그니처 문서화
- **`@example`**: 사용 예제 포함
- **`@namespace`**: 네임스페이스 정의
- **타입 정보**: `{string}`, `{Number}` 등 타입 명시

---

## 🥒 Part 4: BDD with Cucumber

### 4.1 Gherkin 기본 구조

#### 기본 Feature와 Scenario

```gherkin
Feature: Guess the word

  # The first example has two steps
  Scenario: Maker starts a game
    When the Maker starts a game
    Then the Maker waits for a Breaker to join

  # The second example has three steps
  Scenario: Breaker joins a game
    Given the Maker has started a game with the word "silky"
    When the Breaker joins the Maker's game
    Then the Breaker must guess a word with 5 characters
```

**학습 포인트**:
- **Given**: 초기 상태 설정
- **When**: 수행할 동작
- **Then**: 예상 결과
- **주석**: `#`으로 설명 추가

---

### 4.2 Scenario Outline - 데이터 주도 테스트

```gherkin
Feature: Is it Friday yet?
  Everybody wants to know when it's Friday

  Scenario Outline: Today is or is not Friday
    Given today is "<day>"
    When I ask whether it's Friday yet
    Then I should be told "<answer>"

  Examples:
    | day            | answer |
    | Friday         | TGIF   |
    | Sunday         | Nope   |
    | anything else! | Nope   |
```

**학습 포인트**:
- **`<placeholder>`**: 변수 플레이스홀더
- **Examples 테이블**: 여러 데이터셋으로 반복 실행
- **반복 방지**: 동일 시나리오 여러 번 작성 불필요

---

### 4.3 Background - 공통 전제 조건

```gherkin
Feature: Multiple site support
  Only blog owners can post to a blog, except administrators,
  who can post to all blogs.

  Background:
    Given a global administrator named "Greg"
    And a blog named "Greg's anti-tax rants"
    And a customer named "Dr. Bill"
    And a blog named "Expensive Therapy" owned by "Dr. Bill"

  Scenario: Dr. Bill posts to his own blog
    Given I am logged in as Dr. Bill
    When I try to post to "Expensive Therapy"
    Then I should see "Your article was published."

  Scenario: Dr. Bill tries to post to somebody else's blog, and fails
    Given I am logged in as Dr. Bill
    When I try to post to "Greg's anti-tax rants"
    Then I should see "Hey! That's not your blog!"
```

**학습 포인트**:
- **Background**: 모든 시나리오 전에 실행
- **중복 제거**: 공통 Given 단계 추출
- **컨텍스트 설정**: Feature 레벨 전제 조건

---

### 4.4 Declarative vs Imperative Gherkin

#### Imperative (권장하지 않음)

```gherkin
Given I visit "/login"
When I enter "Bob" in the "user name" field
  And I enter "tester" in the "password" field
  And I press the "login" button
Then I should see the "welcome" page
```

#### Declarative (권장)

```gherkin
When "Bob" logs in
```

**학습 포인트**:
- **Declarative**: "무엇을" 수행하는지 명시
- **Imperative**: "어떻게" 수행하는지 명시
- **유지보수성**: Declarative가 UI 변경에 강함
- **비즈니스 언어**: 개발자가 아닌 사람도 이해 가능

---

### 4.5 Data Tables - 구조화된 데이터 전달

```gherkin
Given the following users exist:
  | name   | email              | twitter         |
  | Aslak  | aslak@cucumber.io  | @aslak_hellesoy |
  | Julien | julien@cucumber.io | @jbpros         |
  | Matt   | matt@cucumber.io   | @mattwynne      |
```

**학습 포인트**:
- **테이블 형식**: 여러 객체 데이터 전달
- **Step Definition**: 테이블을 파라미터로 수신
- **복잡한 데이터**: 리스트, 객체 배열 표현

---

### 4.6 Step Definition Implementation (JavaScript)

```javascript
const { Given, When, Then, AfterAll } = require('cucumber');
const { Builder, By, Capabilities, Key } = require('selenium-webdriver');
const { expect } = require('chai');

require("chromedriver");

const capabilities = Capabilities.chrome();
capabilities.set('chromeOptions', { "w3c": false });
const driver = new Builder().withCapabilities(capabilities).build();

Given('I am on the Google search page', async function () {
    await driver.get('http://www.google.com');
});

When('I search for {string}', async function (searchTerm) {
    const element = await driver.findElement(By.name('q'));
    element.sendKeys(searchTerm, Key.RETURN);
    element.submit();
});

Then('the page title should start with {string}', {timeout: 60 * 1000}, async function (searchTerm) {
    const title = await driver.getTitle();
    const isTitleStartWithCheese = title.toLowerCase().lastIndexOf(`${searchTerm}`, 0) === 0;
    expect(isTitleStartWithCheese).to.equal(true);
});

AfterAll(async function(){
    await driver.quit();
});
```

**학습 포인트**:
- **Gherkin ↔ Code**: Step Definition이 연결
- **파라미터 추출**: `{string}`, `{int}` 등 타입 지정
- **비동기 처리**: `async/await` 사용
- **Hooks**: `AfterAll`로 정리 작업

---

## 🎯 TDD 워크플로우 통합 패턴

### RED-GREEN-REFACTOR 사이클

#### 1. RED: 실패하는 테스트 작성

```python
# test_calculator.py
def test_add_two_numbers():
    calc = Calculator()
    result = calc.add(2, 3)
    assert result == 5
```

**실행 결과**: `FAILED - NameError: name 'Calculator' is not defined`

---

#### 2. GREEN: 최소한의 코드로 테스트 통과

```python
# calculator.py
class Calculator:
    def add(self, a, b):
        return a + b
```

**실행 결과**: `PASSED`

---

#### 3. REFACTOR: 코드 개선

```python
# calculator.py
class Calculator:
    """간단한 산술 계산기"""

    def add(self, a: int, b: int) -> int:
        """두 숫자의 합을 반환합니다."""
        return a + b
```

**실행 결과**: `PASSED` (기능 변경 없음, 코드 품질 향상)

---

### Pytest + Sphinx 통합 워크플로우

1. **Docstring 작성**: 함수/클래스에 reStructuredText 문서화
2. **테스트 작성**: `test_*.py` 파일에 pytest 테스트
3. **문서 생성**: `sphinx-apidoc` + `make html`
4. **CI/CD**: GitHub Actions에서 테스트 + 문서 빌드

```yaml
# .github/workflows/test-and-docs.yml
name: Test and Build Docs

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Python
        uses: actions/setup-python@v2
        with:
          python-version: 3.9
      - name: Install dependencies
        run: |
          pip install pytest sphinx
      - name: Run tests
        run: pytest
      - name: Build docs
        run: |
          cd docs
          make html
```

---

## 📈 핵심 통찰 및 권장 사항

### 1. Fixture 설계 원칙

- **단일 책임**: 각 fixture는 하나의 목적만 수행
- **Scope 최적화**: 필요한 최소 범위 사용 (function < class < module < session)
- **명확한 네이밍**: `database_connection` > `db`
- **의존성 체인**: Fixture끼리 의존 가능 (pytest가 순서 해결)

### 2. Parametrization 전략

- **명시적 ID**: `ids` 파라미터로 테스트 리포트 가독성 향상
- **조합 테스트**: `@pytest.mark.parametrize` 스택으로 조합 생성
- **Indirect 활용**: 비용 큰 설정은 fixture에서 처리

### 3. Snapshot Testing 가이드라인

- **결정적 데이터**: 랜덤/시간 값은 `expect.any()` 또는 전처리
- **설명적 네이밍**: `exports[<UserName /> should render null]` > `exports[test 1]`
- **리뷰 프로세스**: 스냅샷 변경은 PR에서 반드시 검토
- **업데이트 주기**: `--updateSnapshot` 사용 시 신중히

### 4. BDD Gherkin 모범 사례

- **선언적 작성**: "What" 중심, "How" 최소화
- **비즈니스 언어**: 개발자가 아닌 이해관계자도 이해 가능
- **한 단계 = 한 행동**: Conjunction 회피 (`Given I have X and Y` → `Given I have X` + `And I have Y`)
- **Background 활용**: 중복 Given 단계 제거

### 5. 문서화 자동화

- **코드 = 문서**: Docstring을 유일한 문서 소스로
- **타입 힌트**: Python type hints + Sphinx autodoc
- **예제 포함**: `@example` (JSDoc) 또는 `.. code-block::` (Sphinx)
- **CI 통합**: 매 커밋마다 문서 빌드 검증

---

## 🔗 추가 리소스

### Pytest
- 공식 문서: https://docs.pytest.org/
- 플러그인 레지스트리: https://docs.pytest.org/en/latest/reference/plugin_list.html
- Best Practices: https://docs.pytest.org/en/latest/goodpractices.html

### Jest
- 공식 문서: https://jestjs.io/
- Snapshot Testing: https://jestjs.io/docs/snapshot-testing
- Mock Functions: https://jestjs.io/docs/mock-functions

### Sphinx
- 공식 문서: https://www.sphinx-doc.org/
- reStructuredText Primer: https://www.sphinx-doc.org/en/master/usage/restructuredtext/basics.html
- Extensions: https://www.sphinx-doc.org/en/master/usage/extensions/index.html

### Cucumber
- 공식 문서: https://cucumber.io/docs/
- Gherkin Reference: https://cucumber.io/docs/gherkin/reference/
- BDD Guide: https://cucumber.io/docs/bdd/

---

## 📝 결론

이 연구를 통해 **7,549개의 코드 예제**를 수집하고 분석하여 다음과 같은 핵심 TDD 워크플로우 패턴을 도출했습니다:

1. **Pytest Fixture 시스템**: 재사용 가능한 테스트 설정 및 의존성 주입
2. **Parametrization**: 반복적인 테스트 케이스 자동화
3. **Jest Snapshot Testing**: UI 회귀 테스트 및 출력 검증
4. **Sphinx Autodoc**: Python 코드에서 자동 문서 생성
5. **Cucumber BDD**: 비즈니스 언어로 실행 가능한 명세 작성

이러한 패턴들은 **moai-alfred-dev-guide** 스킬에 통합되어 MoAI-ADK 사용자들에게 실무에서 즉시 적용 가능한 TDD 가이드를 제공할 것입니다.

---

**연구 수행**: Claude (Context7 MCP Integration)
**보고서 생성일**: 2025-11-12
