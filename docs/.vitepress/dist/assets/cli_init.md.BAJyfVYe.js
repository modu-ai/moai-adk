import{_ as a,c as i,o as n,a3 as l}from"./chunks/framework.L7k57l8l.js";const o=JSON.parse('{"title":"moai init - 프로젝트 초기화","description":"","frontmatter":{},"headers":[],"relativePath":"cli/init.md","filePath":"cli/init.md"}'),p={name:"cli/init.md"};function t(e,s,h,k,d,r){return n(),i("div",null,[...s[0]||(s[0]=[l(`<h1 id="moai-init-프로젝트-초기화" tabindex="-1">moai init - 프로젝트 초기화 <a class="header-anchor" href="#moai-init-프로젝트-초기화" aria-label="Permalink to &quot;moai init - 프로젝트 초기화&quot;">​</a></h1><p><strong>프로젝트를 MoAI-ADK SPEC-First TDD 환경으로 초기화합니다.</strong></p><h2 id="개요" tabindex="-1">개요 <a class="header-anchor" href="#개요" aria-label="Permalink to &quot;개요&quot;">​</a></h2><p><code>moai init</code> 명령은 새 프로젝트를 MoAI-ADK 개발 환경으로 초기화하는 가장 기본적이고 중요한 명령어입니다. 이 명령은 <code>.moai/</code> 및 <code>.claude/</code> 디렉토리 구조를 생성하고, SPEC-First TDD 워크플로우에 필요한 모든 템플릿과 설정 파일을 자동으로 설치합니다.</p><p>초기화 과정은 대화형 위저드를 통해 진행되며, 프로젝트 언어를 자동으로 감지하고 해당 언어에 최적화된 개발 도구 구성을 제안합니다. Commander.js 14.0.1 기반으로 구현되어 안정적이고 사용자 친화적인 경험을 제공합니다.</p><p>초기화는 기존 프로젝트에도 적용할 수 있으며, <code>--force</code> 옵션으로 기존 설정을 덮어쓸 수 있습니다. Personal 모드(로컬 개발)와 Team 모드(GitHub 연동) 중 선택할 수 있어, 개인 프로젝트부터 팀 협업까지 모두 지원합니다.</p><h2 id="기본-구문" tabindex="-1">기본 구문 <a class="header-anchor" href="#기본-구문" aria-label="Permalink to &quot;기본 구문&quot;">​</a></h2><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> [project-name] [options]</span></span></code></pre></div><h3 id="위치-인자" tabindex="-1">위치 인자 <a class="header-anchor" href="#위치-인자" aria-label="Permalink to &quot;위치 인자&quot;">​</a></h3><ul><li><code>project-name</code> (선택): 생성할 프로젝트 이름 <ul><li>생략 시 현재 디렉토리를 초기화</li><li>제공 시 해당 이름의 새 디렉토리 생성</li></ul></li></ul><h3 id="주요-옵션" tabindex="-1">주요 옵션 <a class="header-anchor" href="#주요-옵션" aria-label="Permalink to &quot;주요 옵션&quot;">​</a></h3><table tabindex="0"><thead><tr><th>옵션</th><th>단축</th><th>기본값</th><th>설명</th></tr></thead><tbody><tr><td><code>--template &lt;type&gt;</code></td><td><code>-t</code></td><td><code>standard</code></td><td>사용할 템플릿 (standard, minimal, advanced)</td></tr><tr><td><code>--interactive</code></td><td><code>-i</code></td><td><code>false</code></td><td>대화형 설정 위저드 실행</td></tr><tr><td><code>--backup</code></td><td><code>-b</code></td><td><code>false</code></td><td>설치 전 기존 파일 백업</td></tr><tr><td><code>--force</code></td><td><code>-f</code></td><td><code>false</code></td><td>기존 파일 강제 덮어쓰기</td></tr><tr><td><code>--personal</code></td><td>-</td><td><code>true</code></td><td>Personal 모드로 초기화 (기본값)</td></tr><tr><td><code>--team</code></td><td>-</td><td><code>false</code></td><td>Team 모드로 초기화 (GitHub 연동)</td></tr></tbody></table><h2 id="실제-사용-예제" tabindex="-1">실제 사용 예제 <a class="header-anchor" href="#실제-사용-예제" aria-label="Permalink to &quot;실제 사용 예제&quot;">​</a></h2><h3 id="예제-1-기본-프로젝트-생성" tabindex="-1">예제 1: 기본 프로젝트 생성 <a class="header-anchor" href="#예제-1-기본-프로젝트-생성" aria-label="Permalink to &quot;예제 1: 기본 프로젝트 생성&quot;">​</a></h3><p>가장 일반적인 사용 방법입니다. 새 디렉토리를 만들고 표준 템플릿으로 초기화합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 새 프로젝트 생성</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-awesome-project</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 출력:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 🗿 MoAI-ADK v0.0.1 - Project Initialization</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   Step 1: System Verification</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Node.js  18.19.0</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Git      2.42.0</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ npm      10.2.5</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   Step 2: Configuration</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 📂 Project Name: my-awesome-project</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Detected Language: TypeScript</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 🗿 Mode: Personal</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   Step 3: Directory Structure</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Created .moai/</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Created .claude/</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Created src/</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   Step 4: Template Installation</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Installed 7 agents</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Installed 5 commands</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Installed 8 hooks</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Installed project templates</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Project initialized successfully!</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Next steps:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 1. cd my-awesome-project</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 2. Open in Claude Code</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 3. Run: /alfred:1-spec &quot;Your first feature&quot;</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 프로젝트 디렉토리로 이동</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">cd</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-awesome-project</span></span></code></pre></div><h3 id="예제-2-현재-디렉토리-초기화" tabindex="-1">예제 2: 현재 디렉토리 초기화 <a class="header-anchor" href="#예제-2-현재-디렉토리-초기화" aria-label="Permalink to &quot;예제 2: 현재 디렉토리 초기화&quot;">​</a></h3><p>기존 프로젝트에 MoAI-ADK를 추가할 때 사용합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 기존 프로젝트 디렉토리에서</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">cd</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> existing-project</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 현재 디렉토리 초기화</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 언어 자동 감지 후 해당 언어 설정 적용</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Python 프로젝트라면 pytest, mypy, ruff 추천</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># TypeScript 프로젝트라면 Vitest, Biome 추천</span></span></code></pre></div><h3 id="예제-3-대화형-위저드-사용" tabindex="-1">예제 3: 대화형 위저드 사용 <a class="header-anchor" href="#예제-3-대화형-위저드-사용" aria-label="Permalink to &quot;예제 3: 대화형 위저드 사용&quot;">​</a></h3><p>모든 설정을 단계별로 선택하고 싶을 때 사용합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-project</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --interactive</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 대화형 프롬프트:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 프로젝트 이름: my-project</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 주 개발 언어: (자동 감지됨: TypeScript)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ TypeScript</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Python</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Java</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Go</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Rust</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◉ TypeScript (detected)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 프로젝트 모드:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◉ Personal (로컬 개발)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Team (GitHub 연동)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 템플릿 선택:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◉ Standard (권장)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Minimal (최소 구성)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ◯ Advanced (고급 기능 포함)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 추가 기능:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☑ CI/CD 템플릿</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☑ Docker 설정</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☐ VSCode 설정</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☑ Git hooks</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ 설정 완료! 초기화를 시작합니다...</span></span></code></pre></div><h3 id="예제-4-team-모드-초기화" tabindex="-1">예제 4: Team 모드 초기화 <a class="header-anchor" href="#예제-4-team-모드-초기화" aria-label="Permalink to &quot;예제 4: Team 모드 초기화&quot;">​</a></h3><p>GitHub와 연동하여 팀 협업 환경을 구성합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> team-project</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --team</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Team 모드 추가 설정:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - GitHub repository 연결 확인</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - GitHub CLI (gh) 설치 확인</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - GitHub Actions 워크플로우 생성</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - Issue 템플릿 및 PR 템플릿 설치</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - Team 협업용 훅 설정</span></span></code></pre></div><h3 id="예제-5-기존-설정-강제-덮어쓰기" tabindex="-1">예제 5: 기존 설정 강제 덮어쓰기 <a class="header-anchor" href="#예제-5-기존-설정-강제-덮어쓰기" aria-label="Permalink to &quot;예제 5: 기존 설정 강제 덮어쓰기&quot;">​</a></h3><p>기존 MoAI-ADK 설정을 초기화하거나 업데이트할 때 사용합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 백업 생성 후 강제 덮어쓰기</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --force</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --backup</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 경고 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ⚠️  Warning: Existing .moai/ directory found</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 📦 Creating backup at .moai.backup-2025-01-15-103000/</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 🔄 Overwriting existing configuration...</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Backup created: .moai.backup-2025-01-15-103000/</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ✅ Configuration updated successfully!</span></span></code></pre></div><h3 id="예제-6-minimal-템플릿으로-시작" tabindex="-1">예제 6: Minimal 템플릿으로 시작 <a class="header-anchor" href="#예제-6-minimal-템플릿으로-시작" aria-label="Permalink to &quot;예제 6: Minimal 템플릿으로 시작&quot;">​</a></h3><p>최소한의 구성으로 빠르게 시작하고 싶을 때 사용합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> simple-project</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --template</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> minimal</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Minimal 템플릿 포함 항목:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 필수 에이전트만 (spec-builder, code-builder, doc-syncer)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 기본 명령어만 (/alfred:1-spec, /alfred:2-build, /alfred:3-sync)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 최소 훅 (policy-block, session-notice)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 기본 프로젝트 템플릿</span></span></code></pre></div><h3 id="예제-7-advanced-템플릿으로-시작" tabindex="-1">예제 7: Advanced 템플릿으로 시작 <a class="header-anchor" href="#예제-7-advanced-템플릿으로-시작" aria-label="Permalink to &quot;예제 7: Advanced 템플릿으로 시작&quot;">​</a></h3><p>고급 기능과 추가 도구를 포함한 완전한 환경을 구성합니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> enterprise-project</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --template</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> advanced</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Advanced 템플릿 추가 항목:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 전체 7개 에이전트</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 추가 명령어 (/alfred:8-project)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 전체 8개 훅</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - CI/CD 템플릿 (GitHub Actions, GitLab CI)</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - Docker 및 docker-compose 설정</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 성능 모니터링 도구</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - 보안 스캐닝 설정</span></span></code></pre></div><h2 id="생성되는-디렉토리-구조" tabindex="-1">생성되는 디렉토리 구조 <a class="header-anchor" href="#생성되는-디렉토리-구조" aria-label="Permalink to &quot;생성되는 디렉토리 구조&quot;">​</a></h2><p>초기화 후 생성되는 완전한 프로젝트 구조입니다:</p><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>my-project/</span></span>
<span class="line"><span>├── .moai/                          # MoAI-ADK 설정 루트</span></span>
<span class="line"><span>│   ├── config.json                 # 프로젝트 메인 설정</span></span>
<span class="line"><span>│   │   {</span></span>
<span class="line"><span>│   │     &quot;name&quot;: &quot;my-project&quot;,</span></span>
<span class="line"><span>│   │     &quot;version&quot;: &quot;0.1.0&quot;,</span></span>
<span class="line"><span>│   │     &quot;mode&quot;: &quot;personal&quot;,</span></span>
<span class="line"><span>│   │     &quot;language&quot;: &quot;typescript&quot;,</span></span>
<span class="line"><span>│   │     &quot;created&quot;: &quot;2025-01-15T10:30:00Z&quot;</span></span>
<span class="line"><span>│   │   }</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── memory/                     # 개발 가이드 메모리</span></span>
<span class="line"><span>│   │   └── development-guide.md   # TRUST 5원칙 및 코딩 규칙</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── specs/                      # SPEC 문서 저장소</span></span>
<span class="line"><span>│   │   └── .gitkeep               # Git tracking</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   # TAG는 소스코드에만 존재 (CODE-FIRST)</span></span>
<span class="line"><span>│   # 별도의 tags/ 폴더 불필요 - 코드 직접 스캔</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── project/                    # 프로젝트 메타데이터</span></span>
<span class="line"><span>│   │   ├── product.md             # 제품 정의 (EARS)</span></span>
<span class="line"><span>│   │   ├── structure.md           # 아키텍처 설계</span></span>
<span class="line"><span>│   │   └── tech.md               # 기술 스택 정의</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   └── reports/                   # 동기화 리포트</span></span>
<span class="line"><span>│       └── .gitkeep</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── .claude/                       # Claude Code 통합</span></span>
<span class="line"><span>│   ├── agents/alfred/               # 7개 전문 에이전트</span></span>
<span class="line"><span>│   │   ├── spec-builder.md       # SPEC 작성 전담</span></span>
<span class="line"><span>│   │   ├── code-builder.md       # TDD 구현 전담</span></span>
<span class="line"><span>│   │   ├── doc-syncer.md         # 문서 동기화</span></span>
<span class="line"><span>│   │   ├── cc-manager.md         # Claude Code 설정</span></span>
<span class="line"><span>│   │   ├── debug-helper.md       # 오류 분석</span></span>
<span class="line"><span>│   │   ├── git-manager.md        # Git 작업 자동화</span></span>
<span class="line"><span>│   │   └── trust-checker.md      # 품질 검증</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── commands/alfred/             # 워크플로우 명령어</span></span>
<span class="line"><span>│   │   ├── 8-project.md          # 프로젝트 초기화</span></span>
<span class="line"><span>│   │   ├── 1-spec.md            # SPEC 작성</span></span>
<span class="line"><span>│   │   ├── 2-build.md           # TDD 구현</span></span>
<span class="line"><span>│   │   └── 3-sync.md            # 문서 동기화</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── hooks/alfred/                # 이벤트 훅 (JavaScript)</span></span>
<span class="line"><span>│   │   ├── file-monitor.js       # 파일 변경 감지</span></span>
<span class="line"><span>│   │   ├── language-detector.js  # 언어 자동 감지</span></span>
<span class="line"><span>│   │   ├── policy-block.js       # 보안 정책 강제</span></span>
<span class="line"><span>│   │   ├── pre-write-guard.js    # 쓰기 전 검증</span></span>
<span class="line"><span>│   │   ├── session-notice.js     # 세션 시작 알림</span></span>
<span class="line"><span>│   │   └── steering-guard.js     # 방향성 가이드</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   ├── output-styles/             # 출력 스타일</span></span>
<span class="line"><span>│   │   ├── beginner.md           # 초보자용</span></span>
<span class="line"><span>│   │   ├── study.md             # 학습용</span></span>
<span class="line"><span>│   │   └── pair.md              # 페어 프로그래밍용</span></span>
<span class="line"><span>│   │</span></span>
<span class="line"><span>│   └── settings.json              # Claude Code 설정</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── src/                           # 소스 코드 (언어별)</span></span>
<span class="line"><span>│   └── .gitkeep</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── tests/                         # 테스트 디렉토리</span></span>
<span class="line"><span>│   └── .gitkeep</span></span>
<span class="line"><span>│</span></span>
<span class="line"><span>├── .gitignore                     # Git 제외 파일</span></span>
<span class="line"><span>├── .gitattributes                 # Git 속성</span></span>
<span class="line"><span>└── README.md                      # 프로젝트 README</span></span></code></pre></div><h2 id="템플릿-비교" tabindex="-1">템플릿 비교 <a class="header-anchor" href="#템플릿-비교" aria-label="Permalink to &quot;템플릿 비교&quot;">​</a></h2><p>세 가지 템플릿의 차이점을 이해하고 프로젝트에 맞는 것을 선택하세요.</p><h3 id="standard-템플릿-권장" tabindex="-1">Standard 템플릿 (권장) <a class="header-anchor" href="#standard-템플릿-권장" aria-label="Permalink to &quot;Standard 템플릿 (권장)&quot;">​</a></h3><p><strong>대상</strong>: 대부분의 프로젝트</p><p><strong>포함 항목</strong>:</p><ul><li>✅ 7개 전문 에이전트</li><li>✅ 3단계 워크플로우 명령어</li><li>✅ 6개 핵심 훅</li><li>✅ 3개 출력 스타일</li><li>✅ TRUST 5원칙 개발 가이드</li><li>✅ TAG 시스템 (CODE-FIRST, 소스코드 기반)</li><li>✅ 프로젝트 메타데이터 템플릿</li></ul><p><strong>장점</strong>:</p><ul><li>완전한 SPEC-First TDD 환경</li><li>즉시 사용 가능한 모든 기능</li><li>균형잡힌 구성</li></ul><h3 id="minimal-템플릿" tabindex="-1">Minimal 템플릿 <a class="header-anchor" href="#minimal-템플릿" aria-label="Permalink to &quot;Minimal 템플릿&quot;">​</a></h3><p><strong>대상</strong>: 빠른 프로토타입, 학습용</p><p><strong>포함 항목</strong>:</p><ul><li>✅ 3개 필수 에이전트 (spec-builder, code-builder, doc-syncer)</li><li>✅ 3단계 워크플로우 명령어</li><li>✅ 2개 필수 훅 (policy-block, session-notice)</li><li>✅ 기본 개발 가이드</li><li>✅ 기본 TAG 시스템</li></ul><p><strong>장점</strong>:</p><ul><li>빠른 설치 (&lt; 5초)</li><li>단순한 구조</li><li>학습 곡선 완화</li></ul><p><strong>제한</strong>:</p><ul><li>Git 자동화 없음</li><li>고급 진단 기능 제한</li><li>CI/CD 템플릿 없음</li></ul><h3 id="advanced-템플릿" tabindex="-1">Advanced 템플릿 <a class="header-anchor" href="#advanced-템플릿" aria-label="Permalink to &quot;Advanced 템플릿&quot;">​</a></h3><p><strong>대상</strong>: 엔터프라이즈, 대규모 팀 프로젝트</p><p><strong>포함 항목</strong>:</p><ul><li>✅ Standard 템플릿의 모든 항목</li><li>✅ GitHub Actions 워크플로우</li><li>✅ GitLab CI 설정</li><li>✅ Docker 및 docker-compose</li><li>✅ 성능 모니터링 도구</li><li>✅ 보안 스캐닝 (Snyk, CodeQL)</li><li>✅ API 문서 자동 생성 (TypeDoc, Sphinx)</li><li>✅ 릴리즈 자동화</li></ul><p><strong>장점</strong>:</p><ul><li>프로덕션 준비 완료</li><li>완전 자동화</li><li>엔터프라이즈 기능</li></ul><p><strong>고려사항</strong>:</p><ul><li>복잡한 구성</li><li>추가 도구 필요 (Docker, GitHub CLI)</li><li>설치 시간 증가 (~ 30초)</li></ul><h2 id="personal-vs-team-모드" tabindex="-1">Personal vs Team 모드 <a class="header-anchor" href="#personal-vs-team-모드" aria-label="Permalink to &quot;Personal vs Team 모드&quot;">​</a></h2><h3 id="personal-모드-기본값" tabindex="-1">Personal 모드 (기본값) <a class="header-anchor" href="#personal-모드-기본값" aria-label="Permalink to &quot;Personal 모드 (기본값)&quot;">​</a></h3><p><strong>특징</strong>:</p><ul><li>로컬 Git 저장소 사용</li><li>브랜치 관리는 수동</li><li>GitHub 연동 없음</li><li>혼자 개발하기 최적</li></ul><p><strong>워크플로우</strong>:</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:1-spec</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;New feature&quot;</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 로컬 브랜치 생성 (feature/spec-001-new-feature)</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:2-build</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> SPEC-001</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 로컬에서 TDD 구현</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:3-sync</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 로컬 문서 업데이트</span></span></code></pre></div><h3 id="team-모드" tabindex="-1">Team 모드 <a class="header-anchor" href="#team-모드" aria-label="Permalink to &quot;Team 모드&quot;">​</a></h3><p><strong>특징</strong>:</p><ul><li>GitHub 완전 통합</li><li>Issue 자동 생성</li><li>PR 자동 관리</li><li>협업 워크플로우 최적화</li></ul><p><strong>요구사항</strong>:</p><ul><li>GitHub repository</li><li>GitHub CLI (<code>gh</code>) 설치</li><li>GitHub 인증 완료</li></ul><p><strong>워크플로우</strong>:</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:1-spec</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;New feature&quot;</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → GitHub Issue 생성</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 브랜치 생성 및 연결</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → Draft PR 생성</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:2-build</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> SPEC-001</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → TDD 구현</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 자동 커밋 및 푸시</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:3-sync</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 문서 동기화</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → PR 상태: Draft → Ready for Review</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># → 리뷰어 자동 할당</span></span></code></pre></div><h2 id="출력-결과-해석" tabindex="-1">출력 결과 해석 <a class="header-anchor" href="#출력-결과-해석" aria-label="Permalink to &quot;출력 결과 해석&quot;">​</a></h2><h3 id="성공적인-초기화" tabindex="-1">성공적인 초기화 <a class="header-anchor" href="#성공적인-초기화" aria-label="Permalink to &quot;성공적인 초기화&quot;">​</a></h3><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>✅ Project initialized successfully!</span></span>
<span class="line"><span></span></span>
<span class="line"><span>📂 Project: my-awesome-project</span></span>
<span class="line"><span>📁 Location: /Users/you/projects/my-awesome-project</span></span>
<span class="line"><span>🗿 Mode: Personal</span></span>
<span class="line"><span>🌐 Language: TypeScript</span></span>
<span class="line"><span>📦 Template: Standard</span></span>
<span class="line"><span></span></span>
<span class="line"><span>📊 Installed Components:</span></span>
<span class="line"><span>  ✅ Agents: 7/7</span></span>
<span class="line"><span>  ✅ Commands: 5/5</span></span>
<span class="line"><span>  ✅ Hooks: 8/8</span></span>
<span class="line"><span>  ✅ Templates: ✓</span></span>
<span class="line"><span></span></span>
<span class="line"><span>🚀 Next steps:</span></span>
<span class="line"><span>1. cd my-awesome-project</span></span>
<span class="line"><span>2. Open in Claude Code (VS Code with Claude extension)</span></span>
<span class="line"><span>3. Run system diagnostics: moai doctor</span></span>
<span class="line"><span>4. Start first SPEC: /alfred:1-spec &quot;Your first feature&quot;</span></span>
<span class="line"><span></span></span>
<span class="line"><span>📚 Documentation: https://adk.mo.ai.kr</span></span>
<span class="line"><span>💬 Community: https://mo.ai.kr (오픈 예정)</span></span></code></pre></div><h3 id="경고가-있는-초기화" tabindex="-1">경고가 있는 초기화 <a class="header-anchor" href="#경고가-있는-초기화" aria-label="Permalink to &quot;경고가 있는 초기화&quot;">​</a></h3><div class="language- vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span>⚠️  Warnings detected:</span></span>
<span class="line"><span></span></span>
<span class="line"><span>📦 Existing files found:</span></span>
<span class="line"><span>  - .moai/ (will be skipped, use --force to overwrite)</span></span>
<span class="line"><span>  - .claude/ (will be merged)</span></span>
<span class="line"><span></span></span>
<span class="line"><span>✅ Initialization completed with warnings.</span></span>
<span class="line"><span></span></span>
<span class="line"><span>💡 Recommendations:</span></span>
<span class="line"><span>1. Review existing .moai/config.json</span></span>
<span class="line"><span>2. Backup important files before using --force</span></span>
<span class="line"><span>3. Run &#39;moai doctor&#39; to verify setup</span></span></code></pre></div><h2 id="문제-해결" tabindex="-1">문제 해결 <a class="header-anchor" href="#문제-해결" aria-label="Permalink to &quot;문제 해결&quot;">​</a></h2><h3 id="오류-1-디렉토리가-이미-존재함" tabindex="-1">오류 1: 디렉토리가 이미 존재함 <a class="header-anchor" href="#오류-1-디렉토리가-이미-존재함" aria-label="Permalink to &quot;오류 1: 디렉토리가 이미 존재함&quot;">​</a></h3><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 오류 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">❌</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Error:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Directory</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &#39;my-project&#39;</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> already</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> exists</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 해결 방법:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 방법 1: 다른 이름 사용</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-project-v2</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 방법 2: 기존 디렉토리에서 초기화</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">cd</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-project</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 방법 3: 강제 덮어쓰기 (주의!)</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-project</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --force</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --backup</span></span></code></pre></div><h3 id="오류-2-node-js-버전-불일치" tabindex="-1">오류 2: Node.js 버전 불일치 <a class="header-anchor" href="#오류-2-node-js-버전-불일치" aria-label="Permalink to &quot;오류 2: Node.js 버전 불일치&quot;">​</a></h3><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 오류 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">❌</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Error:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Node.js</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> version</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> 16.x</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> detected</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">Required:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Node.js</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;"> &gt;</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">=</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> 18.0.0</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 해결 방법:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># nvm 사용 시</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">nvm</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> 18</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">nvm</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> use</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> 18</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 또는 공식 사이트에서 다운로드</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># https://nodejs.org</span></span></code></pre></div><h3 id="오류-3-권한-오류-macos-linux" tabindex="-1">오류 3: 권한 오류 (macOS/Linux) <a class="header-anchor" href="#오류-3-권한-오류-macos-linux" aria-label="Permalink to &quot;오류 3: 권한 오류 (macOS/Linux)&quot;">​</a></h3><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 오류 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">❌</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Error:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> EACCES:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> permission</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> denied</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 해결 방법:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 방법 1: npm prefix 변경 (권장)</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">npm</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> config</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> set</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> prefix</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> ~/.npm-global</span></span>
<span class="line"><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">export</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> PATH</span><span style="--shiki-light:#D73A49;--shiki-dark:#F97583;">=~</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">/.npm-global/bin:$PATH</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 방법 2: sudo 사용 (비권장)</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">sudo</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> my-project</span></span></code></pre></div><h3 id="오류-4-git이-설치되지-않음" tabindex="-1">오류 4: Git이 설치되지 않음 <a class="header-anchor" href="#오류-4-git이-설치되지-않음" aria-label="Permalink to &quot;오류 4: Git이 설치되지 않음&quot;">​</a></h3><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 오류 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">❌</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Error:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> not</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> found</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">Git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> is</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> required</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> for</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> MoAI-ADK</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 해결 방법:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># macOS</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">brew</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> git</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Ubuntu/Debian</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">sudo</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> apt-get</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> git</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Windows</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># https://git-scm.com/download/win 에서 다운로드</span></span></code></pre></div><h3 id="오류-5-team-모드-설정-실패" tabindex="-1">오류 5: Team 모드 설정 실패 <a class="header-anchor" href="#오류-5-team-모드-설정-실패" aria-label="Permalink to &quot;오류 5: Team 모드 설정 실패&quot;">​</a></h3><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 오류 메시지:</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">❌</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> Error:</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> GitHub</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> CLI</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> not</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> found</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">Team</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> mode</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> requires</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> GitHub</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> CLI</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;"> (gh)</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 해결 방법:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># GitHub CLI 설치</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># macOS</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">brew</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> gh</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Ubuntu/Debian</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">sudo</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> apt</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> gh</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Windows</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">winget</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> install</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> GitHub.cli</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 인증</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">gh</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> auth</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> login</span></span></code></pre></div><h2 id="고급-사용법" tabindex="-1">고급 사용법 <a class="header-anchor" href="#고급-사용법" aria-label="Permalink to &quot;고급 사용법&quot;">​</a></h2><h3 id="기존-프로젝트-마이그레이션" tabindex="-1">기존 프로젝트 마이그레이션 <a class="header-anchor" href="#기존-프로젝트-마이그레이션" aria-label="Permalink to &quot;기존 프로젝트 마이그레이션&quot;">​</a></h3><p>기존 프로젝트에 MoAI-ADK를 추가하는 단계별 가이드입니다.</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 1. 현재 프로젝트 백업</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> add</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> commit</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -m</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;Backup before MoAI-ADK migration&quot;</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> branch</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> backup-</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">$(</span><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">date</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> +%Y%m%d</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">)</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 2. MoAI-ADK 초기화</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --backup</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 3. 시스템 진단</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> doctor</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 4. 설정 커스터마이징</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">vim</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .moai/config.json</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 5. Git에 추가</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> add</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .moai/</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> .claude/</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">git</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> commit</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> -m</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;Add MoAI-ADK configuration&quot;</span></span></code></pre></div><h3 id="설정-파일-커스터마이징" tabindex="-1">설정 파일 커스터마이징 <a class="header-anchor" href="#설정-파일-커스터마이징" aria-label="Permalink to &quot;설정 파일 커스터마이징&quot;">​</a></h3><p>생성된 <code>config.json</code>을 프로젝트에 맞게 수정할 수 있습니다.</p><div class="language-json vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">json</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">// .moai/config.json</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">{</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;name&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;my-project&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;version&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;0.1.0&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;mode&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;personal&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;language&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;typescript&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;created&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;2025-01-15T10:30:00Z&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">  // 커스터마이징 가능 항목</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;features&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: {</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;autoSync&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">true</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,          </span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">// 자동 문서 동기화</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;strictTDD&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">true</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,          </span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">// 엄격한 TDD 강제</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;coverage&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: {</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">      &quot;threshold&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">85</span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">           // 최소 커버리지 %</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">    }</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">  },</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;tools&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: {</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;testRunner&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;vitest&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,     </span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">// 언어별 자동 감지</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;linter&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;biome&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;formatter&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;">&quot;biome&quot;</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">  },</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">  &quot;git&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: {</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;autoCommit&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">false</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">,        </span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">// 자동 커밋 비활성화</span></span>
<span class="line"><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">    &quot;requireApproval&quot;</span><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">: </span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;">true</span><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">     // 브랜치 생성 시 승인 요구</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">  }</span></span>
<span class="line"><span style="--shiki-light:#24292E;--shiki-dark:#E1E4E8;">}</span></span></code></pre></div><h3 id="다중-언어-프로젝트-설정" tabindex="-1">다중 언어 프로젝트 설정 <a class="header-anchor" href="#다중-언어-프로젝트-설정" aria-label="Permalink to &quot;다중 언어 프로젝트 설정&quot;">​</a></h3><p>프로젝트에서 여러 언어를 사용하는 경우:</p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 1. 주 언어로 초기화</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> init</span><span style="--shiki-light:#005CC5;--shiki-dark:#79B8FF;"> --interactive</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 대화형에서 여러 언어 선택:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 주 개발 언어: TypeScript</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># ? 추가 언어:</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☑ Python</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☑ Go</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;">#   ☐ Java</span></span>
<span class="line"></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># 2. 각 언어별 도구 자동 설정</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - TypeScript: Vitest, Biome</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - Python: pytest, mypy, ruff</span></span>
<span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># - Go: go test, gofmt</span></span></code></pre></div><h2 id="다음-단계" tabindex="-1">다음 단계 <a class="header-anchor" href="#다음-단계" aria-label="Permalink to &quot;다음 단계&quot;">​</a></h2><p>초기화가 완료되면 다음 작업을 진행하세요:</p><ol><li><p><strong>시스템 진단 실행</strong></p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> doctor</span></span></code></pre></div><p>→ <a href="/cli/doctor.html">moai doctor 가이드</a> 참조</p></li><li><p><strong>프로젝트 상태 확인</strong></p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">moai</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> status</span></span></code></pre></div><p>→ <a href="/cli/status.html">moai status 가이드</a> 참조</p></li><li><p><strong>첫 SPEC 작성</strong></p><div class="language-bash vp-adaptive-theme"><button title="Copy Code" class="copy"></button><span class="lang">bash</span><pre class="shiki shiki-themes github-light github-dark vp-code" tabindex="0"><code><span class="line"><span style="--shiki-light:#6A737D;--shiki-dark:#6A737D;"># Claude Code에서</span></span>
<span class="line"><span style="--shiki-light:#6F42C1;--shiki-dark:#B392F0;">/alfred:1-spec</span><span style="--shiki-light:#032F62;--shiki-dark:#9ECBFF;"> &quot;사용자 인증 기능&quot;</span></span></code></pre></div><p>→ <a href="/guide/spec-first-tdd.html">SPEC-First TDD 가이드</a> 참조</p></li><li><p><strong>3단계 워크플로우 학습</strong> → <a href="/guide/workflow.html">3단계 워크플로우 완전 가이드</a> 참조</p></li></ol><h2 id="관련-문서" tabindex="-1">관련 문서 <a class="header-anchor" href="#관련-문서" aria-label="Permalink to &quot;관련 문서&quot;">​</a></h2><ul><li><a href="/getting-started/quick-start.html">빠른 시작 가이드</a> - 5분 안에 시작하기</li><li><a href="/cli/doctor.html">moai doctor</a> - 시스템 진단</li><li><a href="/getting-started/installation.html">설치 가이드</a> - 상세 설치 방법</li><li><a href="/guide/workflow.html">3단계 워크플로우</a> - SPEC → Build → Sync</li></ul><hr><p><strong>다음 읽기</strong>: <a href="/cli/doctor.html">moai doctor - 시스템 진단</a></p>`,107)])])}const g=a(p,[["render",t]]);export{o as __pageData,g as default};
