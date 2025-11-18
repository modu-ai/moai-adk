---
name: moai-domain-figma
description: "Figma API, MCP 통합, Design-to-Code 워크플로우를 위한 종합 Domain Expertise Skill. REST API v1, Variables API (Design Tokens), Code Connect, Dev Mode 통합, Plugin 개발까지 완벽 커버"
tier: medium-advanced
version: "1.0.0"
updated: 2025-11-16
status: stable
tags: [figma, design-systems, design-tokens, mcp, code-connect, dev-mode, plugin-api, variables-api, dtcg]
allowed-tools: [Read, Write, WebFetch, Bash]
related-skills:
  - moai-design-systems
  - moai-lang-typescript
  - moai-domain-frontend
  - moai-essentials-perf
---

# Figma Domain Expertise Skill - Enterprise Design System Integration

**Purpose**: 제공하는 완벽한 Figma 통합 지식으로, REST API, Variables API, Code Connect, Plugin 개발, Dev Mode를 활용하여 Design-to-Code 워크플로우를 구현하고 Design System을 구축합니다.

**When to use this Skill**:
- Figma REST API v1 통합 (파일, 컴포넌트, 변수, 이미지 조회)
- Variables API로 Design Tokens (DTCG) 관리
- Code Connect로 Design-to-Code 자동화 (React, Vue, SwiftUI, Compose)
- Figma Plugin 개발
- Dev Mode MCP 통합
- Design System 구축 및 라이브러리 관리
- Context7 MCP로 최신 API 정보 자동 조회

**Latest Standards** (November 2025):
- **Figma REST API**: v1 (2025년 최신)
- **Variables API**: Enterprise 지원, 5,000개 변수/컬렉션, 40개 모드
- **DTCG Specification**: 2025.10 (Design Tokens 표준)
- **Code Connect**: React, Vue, SwiftUI, Jetpack Compose 지원
- **Dev Mode MCP**: Image/SVG assets, component properties
- **Webhooks**: v2 (FILE_UPDATE, LIBRARY_UPDATE events)

---

## Progressive Disclosure Structure

### Level 1: Quick Start Overview (Read This First) - 500 words

**Figma 생태계의 3가지 핵심 요소**:

#### 1. Figma REST API (파일, 컴포넌트, 데이터 조회)

Figma REST API는 디자인 파일의 메타데이터, 컴포넌트, 이미지를 프로그래매틱하게 접근합니다.

**주요 기능**:
- 파일 정보 조회 (레이아웃, 컴포넌트, 스타일)
- 컴포넌트 목록 및 메타데이터
- 이미지 렌더링 (PNG, SVG, PDF)
- 변수 (Variables) 조회 및 수정 (Enterprise)
- 라이브러리 분석

**인증 방법** (3가지):
1. Personal Access Token: `X-Figma-Token` 헤더
2. OAuth 2.0: client_id, client_secret, authorization code
3. SCIM API (Enterprise): Bearer token

**사용 권한**:
- `files:read` - 모든 파일 읽기 (기본)
- `file_variables:read` - Variables 읽기 (Enterprise)
- `file_variables:write` - Variables 쓰기 (Enterprise)
- `library_analytics:read` - 라이브러리 분석 (Enterprise)
- `org:activity_log_read` - 활동 로그 (Enterprise)

#### 2. Variables API (Design Tokens, DTCG 표준)

Variables API는 Design Tokens를 프로그래매틱하게 관리하는 Enterprise 기능입니다. W3C DTCG 2025.10 표준을 따릅니다.

**4가지 변수 타입**:
- `BOOLEAN` - true/false 값
- `FLOAT` - 숫자 (간격, 크기, 불투명도)
- `STRING` - 텍스트 (폰트 이름, 간격 키)
- `COLOR` - 색상 (RGB 객체)

**핵심 개념**:
1. **Variable Collection** - 변수의 그룹 (예: Semantic Colors)
2. **Modes** - 컬렉션 내 서로 다른 변수 값 조합 (Light, Dark, HighContrast)
3. **Variable** - 각 모드별로 다른 값을 가질 수 있는 토큰
4. **Code Syntax** - WEB, ANDROID, iOS 플랫폼별 코드 생성

**예시**:
```
Collection: Semantic Colors
├── Mode: Light
│   ├── color.text.primary = #000000
│   ├── color.text.secondary = #666666
│   └── color.background = #FFFFFF
└── Mode: Dark
    ├── color.text.primary = #FFFFFF
    ├── color.text.secondary = #CCCCCC
    └── color.background = #1A1A1A
```

#### 3. Code Connect (Design-to-Code, 자동화)

Code Connect는 Figma 컴포넌트를 코드베이스의 실제 컴포넌트와 연결하여, Design과 Code가 항상 동기화되도록 합니다.

**지원 프레임워크**:
- React + TypeScript (가장 강력)
- Vue 3
- SwiftUI (iOS)
- Jetpack Compose (Android)
- HTML/CSS (기초)

**매핑 방식**:
- Figma Component Variant → Props
- Component Property → TypeScript 타입
- 이미지 에셋 → Import 문
- 텍스트 → 컴포넌트 children

#### 4. Design System 아키텍처 (Atomic Design)

Figma에서 Design System을 체계적으로 구축하는 방법:

**계층 구조**:
1. **Foundation** - Colors, Typography, Spacing, Effects (Variables)
2. **Components** - Atoms (Button, Input), Molecules (FormInput), Organisms (Navigation)
3. **Patterns** - Templates (Page layouts)
4. **Documentation** - Code Connect definitions

**Variable Collections 구조**:
- **Primitive Colors** - 기본 색상 (blue-500, gray-700)
- **Semantic Colors** - 의미론적 색상 (text-primary, background-surface)
- **Typography** - 폰트 크기, 두께, 라인높이
- **Spacing** - 마진, 패딩, 간격 (8px 기반 단위)
- **Effects** - 그림자, 테두리, 불투명도

#### 5. MCP 도구 (4개 핵심)

Figma Dev Mode MCP 서버는 4개 주요 도구를 제공합니다:

**Tool 1: get_design_context** (가장 중요!)
- Figma 디자인에서 React/Vue/HTML 코드 직접 추출
- 입력: fileKey, nodeId, clientLanguages, clientFrameworks
- 출력: 완성된 컴포넌트 코드 + CSS

**Tool 2: get_variable_defs**
- Design Tokens 정의 조회
- 입력: fileKey, nodeId
- 출력: DTCG JSON 형식 변수 정의

**Tool 3: get_screenshot**
- Figma 디자인의 시각적 미리보기
- PNG/SVG 형식
- 원본 디자인과 비교 검증용

**Tool 4: get_metadata**
- 전체 페이지/프레임 구조
- XML 형식 (노드 ID, 레이어 이름, 위치, 크기)
- 컴포넌트 계층 분석용

#### 6. 의사결정 체크리스트

**Figma 통합 시작 전 확인사항**:
- [ ] Figma Token 준비 (Personal Access Token 또는 OAuth)
- [ ] 프로젝트의 Design System Figma 파일 식별
- [ ] 목표 프레임워크 결정 (React, Vue, SwiftUI, etc)
- [ ] Design Token 이름 규칙 정의 (예: color.text.primary)
- [ ] Component Variant 구조 정의
- [ ] 팀 협업 방식 결정 (Design-first vs Code-first)
- [ ] Multi-mode 필요성 검토 (Light/Dark/HighContrast)

---

### Level 2: Implementation Patterns - 1,500 words

#### Pattern 1: REST API 통합 (파일 및 컴포넌트 데이터 조회)

**Figma REST API 인증 설정**:

```typescript
// Personal Access Token 방식 (가장 간단)
const FIGMA_TOKEN = process.env.FIGMA_TOKEN // figd_xxxxx...
const fileKey = "ABC123XYZ"

const response = await fetch(
  `https://api.figma.com/v1/files/${fileKey}`,
  {
    headers: {
      'X-Figma-Token': FIGMA_TOKEN,
      'Content-Type': 'application/json'
    }
  }
)

const data = await response.json()
// data.document = 전체 디자인 구조
// data.components = 컴포넌트 목록
// data.styles = 스타일 목록
```

**주요 REST API 엔드포인트**:

1. **GET /v1/files/{fileKey}** - 파일 정보 조회
   - 매개변수: version, ids, depth, geometry, plugin_data, branch_data
   - 반환: 전체 파일 데이터 (노드, 컴포넌트, 스타일)

2. **GET /v1/files/{fileKey}/variables/local** - 로컬 변수 조회
   - 매개변수: 없음
   - 반환: Variable, VariableCollection, Mode 정보
   - 권한: file_variables:read (Enterprise)

3. **GET /v1/files/{fileKey}/components** - 컴포넌트 목록
   - 매개변수: page_size, before, after
   - 반환: 컴포넌트 메타데이터, 키, 설명

4. **GET /v1/images/{fileKey}** - 이미지 렌더링
   - 매개변수: ids (노드 ID 리스트), scale, format, use_absolute_bounds, svg_outline_text
   - 반환: { images: { [nodeId]: "https://..." } }
   - format: "png", "jpg", "svg", "pdf"

5. **POST /v1/variables** - 변수 값 수정
   - 매개변수: file_key, library_file_key, variables (수정할 변수 목록)
   - 반환: 수정된 변수 정보
   - 권한: file_variables:write (Enterprise)

6. **GET /v1/dev_resources** - Dev Resources 조회
   - 매개변수: file_key, node_ids
   - 반환: Dev Resource 목록 (문서, 참고 링크)

7. **GET /v1/analytics/libraries/{library_file_key}/variable/usages** - 변수 사용 분석
   - 매개변수: 없음
   - 반환: 어느 파일에서 어느 변수를 사용하는지 분석
   - 권한: library_analytics:read (Enterprise)

**TypeScript로 REST API 호출**:

```typescript
interface GetFileParams {
  version?: string
  ids?: string[]
  depth?: number
  geometry?: boolean
  plugin_data?: boolean
  branch_data?: boolean
}

interface Variable {
  id: string
  name: string
  variableCollectionId: string
  resolvedType: 'BOOLEAN' | 'FLOAT' | 'STRING' | 'COLOR'
  valuesByMode: Record<string, VariableValue>
  scopes?: VariableScope[]
  codeSyntax?: Record<string, string>
  description?: string
}

interface Component {
  key: string
  file_key: string
  node_id: string
  thumbnail_url: string
  name: string
  description: string
  created_at: string
  updated_at: string
  user: { id: string; handle: string; email?: string }
  remote: boolean
}

async function getFileData(fileKey: string, params: GetFileParams = {}) {
  const queryParams = new URLSearchParams()
  if (params.version) queryParams.append('version', params.version)
  if (params.ids) queryParams.append('ids', params.ids.join(','))
  if (params.depth !== undefined) queryParams.append('depth', params.depth.toString())
  if (params.geometry !== undefined) queryParams.append('geometry', params.geometry.toString())

  const response = await fetch(
    `https://api.figma.com/v1/files/${fileKey}?${queryParams}`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_TOKEN }
    }
  )
  return response.json()
}

async function getVariables(fileKey: string) {
  const response = await fetch(
    `https://api.figma.com/v1/files/${fileKey}/variables/local`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_TOKEN }
    }
  )
  return response.json() // { variables: Variable[], variableCollections: VariableCollection[] }
}

async function getComponents(fileKey: string) {
  const response = await fetch(
    `https://api.figma.com/v1/files/${fileKey}/components`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_TOKEN }
    }
  )
  return response.json() // { components: Component[] }
}

async function renderImages(fileKey: string, nodeIds: string[], scale = 1, format = 'png') {
  const response = await fetch(
    `https://api.figma.com/v1/images/${fileKey}?ids=${nodeIds.join(',')}&scale=${scale}&format=${format}`,
    {
      headers: { 'X-Figma-Token': process.env.FIGMA_TOKEN }
    }
  )
  return response.json() // { images: { [nodeId]: "https://..." } }
}
```

#### Pattern 2: Variables API (Design Tokens, DTCG 표준)

**변수 타입과 값**:

```typescript
// BOOLEAN 변수
type VariableValue = boolean | number | string | RGBA

interface RGBA {
  r: number // 0-1
  g: number // 0-1
  b: number // 0-1
  a: number // 0-1 (알파, 불투명도)
}

// 변수 컬렉션 예시: Semantic Colors
const variableCollection = {
  id: "VariableCollection:1:2",
  name: "Semantic Colors",
  modes: [
    { modeId: "Mode:1:2", name: "Light" },
    { modeId: "Mode:2:3", name: "Dark" },
    { modeId: "Mode:3:4", name: "HighContrast" }
  ],
  defaultModeId: "Mode:1:2",
  variableIds: ["Variable:4:5", "Variable:5:6", "Variable:6:7"],
  hiddenFromPublishing: false,
  remote: false
}

// 변수 정의 예시: text-primary 색상
const variable = {
  id: "Variable:4:5",
  name: "color/text/primary",
  variableCollectionId: "VariableCollection:1:2",
  resolvedType: "COLOR",
  valuesByMode: {
    "Mode:1:2": { r: 0, g: 0, b: 0, a: 1 },           // Light: 검은색
    "Mode:2:3": { r: 1, g: 1, b: 1, a: 1 },           // Dark: 흰색
    "Mode:3:4": { r: 0, g: 0, b: 0, a: 1 }            // HighContrast: 검은색
  },
  scopes: ["ALL_FILLS", "ALL_STROKES"],
  codeSyntax: {
    WEB: "--color-text-primary",
    ANDROID: "colorTextPrimary",
    iOS: "ColorTextPrimary"
  },
  description: "Primary text color for light/dark/high contrast modes"
}

// 변수 바인딩 (Component에서 변수 사용)
interface BoundVariable {
  type: "VARIABLE_ALIAS",
  id: string // Variable ID
}

interface ComponentNode {
  // ... other properties
  boundVariables: {
    fills?: BoundVariable[]           // 배경 색상
    strokes?: BoundVariable[]          // 테두리 색상
    effects?: BoundVariable[]          // 그림자, 블러
    opacity?: BoundVariable            // 불투명도
    strokeWeight?: BoundVariable       // 테두리 두께
    componentProperties?: {
      [propertyName: string]: BoundVariable | string | number | boolean
    }
  }
}
```

**변수 생성 및 수정 (Plugin API)**:

```typescript
// Plugin API로 변수 생성 (Figma 플러그인 내에서만 가능)

// 컬렉션 생성
const collection = figma.variables.createVariableCollection("Semantic Colors")

// 모드 추가
const lightModeId = collection.modes[0].modeId
const darkModeId = collection.addMode("dark")

// 변수 생성
const colorVar = figma.variables.createVariable("color/text/primary", collection, "COLOR")

// 모드별 값 설정
colorVar.setValueForMode(lightModeId, { r: 0, g: 0, b: 0, a: 1 })
colorVar.setValueForMode(darkModeId, { r: 1, g: 1, b: 1, a: 1 })

// 코드 신택스 설정 (플랫폼별)
colorVar.setVariableCodeSyntax("WEB", "--color-text-primary")
colorVar.setVariableCodeSyntax("ANDROID", "colorTextPrimary")
colorVar.setVariableCodeSyntax("iOS", "ColorTextPrimary")

// 변수 설명
colorVar.description = "Primary text color"

// REST API로 변수 수정 (Enterprise Token 필요)
// POST /v1/variables
const updateResponse = await fetch(
  'https://api.figma.com/v1/variables',
  {
    method: 'POST',
    headers: {
      'X-Figma-Token': ENTERPRISE_TOKEN,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      file_key: fileKey,
      variables: [
        {
          id: "Variable:4:5",
          // valuesByMode 업데이트
          values_by_mode: {
            "Mode:1:2": { r: 0.2, g: 0.4, b: 0.8, a: 1 }  // 파란색으로 변경
          }
        }
      ]
    })
  }
)
```

**DTCG JSON 형식 (W3C 표준)**:

```json
{
  "$schema": "https://tr.designtokens.org/format/",
  "$tokens": {
    "color": {
      "$type": "color",
      "primary": {
        "50": { "$value": "#f0f9ff" },
        "100": { "$value": "#dbeafe" },
        "500": { "$value": "#0ea5e9" },
        "900": { "$value": "#0c2d4a" }
      },
      "semantic": {
        "text": {
          "primary": {
            "$value": "{color.gray.900}",
            "$description": "Primary text color - links to gray-900"
          },
          "secondary": {
            "$value": "{color.gray.600}",
            "$description": "Secondary text color"
          },
          "disabled": {
            "$value": "{color.gray.400}",
            "$description": "Disabled text color"
          }
        },
        "background": {
          "surface": {
            "$value": "#ffffff",
            "$description": "Default background"
          },
          "elevated": {
            "$value": "#f9fafb",
            "$description": "Elevated surface (cards, modals)"
          }
        }
      }
    },
    "spacing": {
      "$type": "dimension",
      "xs": { "$value": "4px" },
      "sm": { "$value": "8px" },
      "md": { "$value": "16px" },
      "lg": { "$value": "24px" },
      "xl": { "$value": "32px" }
    },
    "typography": {
      "heading": {
        "$type": "typography",
        "1": {
          "$value": {
            "fontFamily": "Inter",
            "fontSize": "32px",
            "fontWeight": "700",
            "lineHeight": "1.25"
          }
        },
        "2": {
          "$value": {
            "fontFamily": "Inter",
            "fontSize": "24px",
            "fontWeight": "700",
            "lineHeight": "1.33"
          }
        }
      },
      "body": {
        "$type": "typography",
        "default": {
          "$value": {
            "fontFamily": "Inter",
            "fontSize": "16px",
            "fontWeight": "400",
            "lineHeight": "1.5"
          }
        }
      }
    },
    "effect": {
      "shadow": {
        "$type": "shadow",
        "sm": {
          "$value": {
            "offsetX": "0px",
            "offsetY": "1px",
            "blur": "2px",
            "color": "#00000014"
          }
        },
        "lg": {
          "$value": {
            "offsetX": "0px",
            "offsetY": "10px",
            "blur": "15px",
            "color": "#0000001a"
          }
        }
      }
    }
  },
  "$modes": {
    "Light": {
      "color.semantic.text.primary": { "$value": "{color.gray.900}" },
      "color.semantic.text.secondary": { "$value": "{color.gray.600}" }
    },
    "Dark": {
      "color.semantic.text.primary": { "$value": "{color.gray.50}" },
      "color.semantic.text.secondary": { "$value": "{color.gray.300}" }
    }
  }
}
```

#### Pattern 3: Code Connect 워크플로우 (Design-to-Code)

**React/TypeScript 예시**:

```typescript
// src/components/Button.tsx
import React from 'react'
import figma from '@figma/code-connect'

interface ButtonProps {
  variant: 'primary' | 'secondary' | 'destructive'
  size: 'sm' | 'md' | 'lg'
  disabled?: boolean
  icon?: React.ReactNode
  label: string
  onClick?: () => void
}

export const Button: React.FC<ButtonProps> = ({
  variant = 'primary',
  size = 'md',
  disabled = false,
  icon,
  label,
  onClick
}) => {
  const baseClasses = 'font-medium rounded transition-colors'

  const variantClasses = {
    primary: 'bg-blue-600 text-white hover:bg-blue-700 disabled:bg-gray-400',
    secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 disabled:bg-gray-100',
    destructive: 'bg-red-600 text-white hover:bg-red-700 disabled:bg-gray-400'
  }

  const sizeClasses = {
    sm: 'px-3 py-1 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg'
  }

  return (
    <button
      className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses[size]}`}
      disabled={disabled}
      onClick={onClick}
    >
      {icon && <span className="mr-2">{icon}</span>}
      {label}
    </button>
  )
}

// Figma Code Connect 정의
figma.connect(
  Button,
  'https://figma.com/file/ABC123XYZ/DesignSystem?node-id=1:2',
  {
    variant: {
      'Variant 1': 'primary',
      'Variant 2': 'secondary',
      'Variant 3': 'destructive'
    },
    props: {
      variant: figma.enum('Variant', {
        'Variant 1': 'primary',
        'Variant 2': 'secondary',
        'Variant 3': 'destructive'
      }),
      size: figma.enum('Size', {
        'Small': 'sm',
        'Medium': 'md',
        'Large': 'lg'
      }),
      disabled: figma.boolean('Disabled'),
      icon: figma.instance('Icon'),  // nested component
      label: figma.string('Label'),
      onClick: figma.boolean('Disabled') ? undefined : () => alert('clicked')
    },
    example: (props) => (
      <Button
        variant={props.variant}
        size={props.size}
        disabled={props.disabled}
        icon={props.icon}
        label={props.label}
        onClick={props.onClick}
      />
    )
  }
)

export default Button
```

**Vue 3 예시**:

```vue
<template>
  <button
    :class="[baseClasses, variantClasses, sizeClasses]"
    :disabled="disabled"
    @click="$emit('click')"
  >
    <span v-if="icon" class="mr-2"><component :is="icon" /></span>
    {{ label }}
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import figma from '@figma/code-connect'

interface Props {
  variant?: 'primary' | 'secondary' | 'destructive'
  size?: 'sm' | 'md' | 'lg'
  disabled?: boolean
  icon?: any
  label: string
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  disabled: false
})

const baseClasses = computed(() => 'font-medium rounded transition-colors')

const variantClasses = computed(() => ({
  primary: 'bg-blue-600 text-white hover:bg-blue-700 disabled:bg-gray-400',
  secondary: 'bg-gray-200 text-gray-900 hover:bg-gray-300 disabled:bg-gray-100',
  destructive: 'bg-red-600 text-white hover:bg-red-700 disabled:bg-gray-400'
}[props.variant]))

const sizeClasses = computed(() => ({
  sm: 'px-3 py-1 text-sm',
  md: 'px-4 py-2 text-base',
  lg: 'px-6 py-3 text-lg'
}[props.size]))

// Vue Code Connect
figma.connect(
  'Button',
  'https://figma.com/file/ABC123XYZ/DesignSystem?node-id=1:2',
  {
    props: {
      variant: figma.enum('Variant', {
        'Variant 1': 'primary',
        'Variant 2': 'secondary',
        'Variant 3': 'destructive'
      }),
      size: figma.enum('Size', {
        'Small': 'sm',
        'Medium': 'md',
        'Large': 'lg'
      }),
      disabled: figma.boolean('Disabled'),
      label: figma.string('Label')
    }
  }
)
</script>
```

**SwiftUI 예시**:

```swift
import SwiftUI
import Figma

struct ButtonDocument: FigmaConnect {
  let component = Button.self
  let figmaNodeUrl = "https://figma.com/file/ABC123XYZ/DesignSystem?node-id=1:2"

  @FigmaEnum("Variant")
  var variant: ButtonStyle = .primary

  @FigmaEnum("Size")
  var size: ButtonSize = .medium

  @FigmaBoolean("Disabled")
  var disabled: Bool = false

  @FigmaString("Label")
  var label: String = "Button"

  var body: some View {
    Button(action: {}) {
      Text(label)
    }
    .buttonStyle(style)
    .disabled(disabled)
  }

  var style: ButtonStyle {
    switch (variant, size) {
    case (.primary, .small):
      return .primary(size: .small)
    case (.primary, .medium):
      return .primary(size: .medium)
    case (.primary, .large):
      return .primary(size: .large)
    case (.secondary, _):
      return .secondary(size: size)
    case (.destructive, _):
      return .destructive(size: size)
    }
  }
}

enum ButtonStyle {
  case primary(size: ButtonSize)
  case secondary(size: ButtonSize)
  case destructive(size: ButtonSize)
}

enum ButtonSize {
  case small, medium, large
}

struct Button: View {
  let variant: ButtonStyle
  let disabled: Bool
  let label: String
  let action: () -> Void

  var body: some View {
    Button(action: action) {
      Text(label)
        .font(.system(size: fontSize, weight: .medium))
        .foregroundColor(foregroundColor)
        .padding(.horizontal, paddingX)
        .padding(.vertical, paddingY)
        .background(backgroundColor)
        .cornerRadius(8)
        .disabled(disabled)
    }
  }

  var fontSize: CGFloat {
    switch variant {
    case .primary(let size), .secondary(let size), .destructive(let size):
      switch size {
      case .small: return 12
      case .medium: return 16
      case .large: return 18
      }
    }
  }

  // ... more properties
}
```

**Jetpack Compose 예시**:

```kotlin
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.runtime.Composable
import com.figma.code.connect.*

@FigmaConnect(url = "https://figma.com/file/ABC123XYZ/DesignSystem?node-id=1:2")
@Composable
fun ButtonDocument() {
  @FigmaEnum("Variant", mapOf(
    "Variant 1" to "primary",
    "Variant 2" to "secondary",
    "Variant 3" to "destructive"
  ))
  val variant = "primary"

  @FigmaEnum("Size", mapOf(
    "Small" to "sm",
    "Medium" to "md",
    "Large" to "lg"
  ))
  val size = "md"

  @FigmaBoolean("Disabled")
  val disabled = false

  @FigmaString("Label")
  val label = "Button"

  val colors = when (variant) {
    "primary" -> ButtonDefaults.buttonColors(
      containerColor = Color(0xFF0EA5E9),
      contentColor = Color.White
    )
    "secondary" -> ButtonDefaults.buttonColors(
      containerColor = Color(0xFFE5E7EB),
      contentColor = Color(0xFF111827)
    )
    "destructive" -> ButtonDefaults.buttonColors(
      containerColor = Color(0xFFEF4444),
      contentColor = Color.White
    )
    else -> ButtonDefaults.buttonColors()
  }

  val (paddingH, paddingV, fontSize) = when (size) {
    "sm" -> Triple(12.dp, 8.dp, 12.sp)
    "md" -> Triple(16.dp, 12.dp, 16.sp)
    "lg" -> Triple(24.dp, 16.dp, 18.sp)
    else -> Triple(16.dp, 12.dp, 16.sp)
  }

  Button(
    onClick = {},
    enabled = !disabled,
    colors = colors,
    modifier = Modifier.padding(horizontal = paddingH, vertical = paddingV)
  ) {
    Text(
      text = label,
      fontSize = fontSize,
      fontWeight = FontWeight.Medium
    )
  }
}
```

#### Pattern 4: Dev Mode 활용 (MCP 서버와 통합)

**Dev Resources API**:

```typescript
// Figma Plugin API로 Dev Resources 관리

// Dev Resources 추가
const node = figma.currentPage.selection[0]

if (node && 'addDevResourceAsync' in node) {
  await node.addDevResourceAsync({
    name: "Component Documentation",
    url: "https://example.com/docs/button"
  })

  await node.addDevResourceAsync({
    name: "Storybook",
    url: "https://storybook.example.com/?path=/story/button"
  })

  await node.addDevResourceAsync({
    name: "GitHub Source",
    url: "https://github.com/example/components/blob/main/Button.tsx"
  })
}

// Dev Resources 조회
const resources = await node.getDevResourcesAsync()
// resources = [
//   { key: 'dev-1', name: 'Component Documentation', url: 'https://...' },
//   { key: 'dev-2', name: 'Storybook', url: 'https://...' },
//   { key: 'dev-3', name: 'GitHub Source', url: 'https://...' }
// ]

// Dev Resources 수정
await node.editDevResourceAsync('dev-1', {
  url: "https://updated-docs.example.com/button"
})

// Dev Resources 삭제
await node.deleteDevResourceAsync('dev-1')
```

**CSS 생성 (Dev Mode)**:

```typescript
// Dev Mode에서 선택한 element의 CSS 자동 생성
const node = figma.currentPage.selection[0]

if (node && 'getCSSAsync' in node) {
  const css = await node.getCSSAsync()
  console.log(css)
  // 출력 예:
  // .button {
  //   font-size: 16px;
  //   font-weight: 700;
  //   color: #ffffff;
  //   background-color: #0ea5e9;
  //   padding: 8px 16px;
  //   border-radius: 4px;
  //   border: none;
  //   cursor: pointer;
  // }
}
```

**Asset 추출**:

```typescript
// 이미지/SVG를 파일로 Export

const node = figma.currentPage.selection[0]

// PNG로 Export (2배 배율)
const pngBytes = await node.exportAsync({
  format: 'PNG',
  constraint: { type: 'SCALE', value: 2 }
})

// SVG로 Export (텍스트 아웃라인)
const svgBytes = await node.exportAsync({
  format: 'SVG',
  svgOutlineText: true,
  svgIdAttribute: true
})

// PDF로 Export
const pdfBytes = await node.exportAsync({
  format: 'PDF'
})
```

**Asset 감지 (Dev Mode)**:

```typescript
// Dev Mode에서 자동으로 아이콘/이미지 감지
const node = figma.currentPage.selection[0]

if ('isAsset' in node && node.isAsset) {
  console.log("이 노드는 재사용 가능한 에셋입니다")

  // 에셋으로 자동 분류 (자동화)
  // → Component Library에 추가 제안
  // → SVG로 Export 자동화
  // → Image optimization 자동화
}
```

#### Pattern 5: Design-to-Code 자동화 (전체 워크플로우)

**Step-by-step Design → React 변환**:

```typescript
// 1단계: Figma URL 파싱
const figmaUrl = 'https://figma.com/design/ABC123XYZ/LoginPage?node-id=10-25'
const urlMatch = figmaUrl.match(/design\/([^/]+)\/[^?]*\?node-id=(\d+)-(\d+)/)
const fileKey = urlMatch[1]
const nodeId = `${urlMatch[2]}:${urlMatch[3]}`

// 2단계: MCP로 Design Context 추출
const designContext = await mcp_figma_get_design_context({
  fileKey,
  nodeId,
  clientLanguages: 'typescript',
  clientFrameworks: 'react'
})
// 결과: React 컴포넌트 코드 + CSS + 이미지 URL

// 3단계: Design Tokens 추출
const tokenDefs = await mcp_figma_get_variable_defs({
  fileKey,
  nodeId
})
// 결과: DTCG JSON 형식 변수 정의

// 4단계: 추출한 코드 검증
// - TypeScript 타입 추가
// - Props 인터페이스 생성
// - Storybook 메타데이터 추가

// 5단계: Design Tokens 변환
// - CSS Variables 생성
// - Tailwind Config 생성
// - TypeScript 타입 생성

// 6단계: 최종 파일 생성
// - src/components/LoginPage.tsx (React)
// - src/styles/tokens.json (Design Tokens)
// - src/components/LoginPage.stories.tsx (Storybook)
// - src/components/LoginPage.test.tsx (Unit Test)
```

**실제 코드 예시**:

```typescript
// Design Context에서 추출한 React 코드 (MCP가 생성함)
export function LoginForm() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)
    try {
      // 로그인 로직
      await login(email, password)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-2">
        <label htmlFor="email" className="block text-sm font-medium text-gray-700">
          이메일
        </label>
        <input
          id="email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
          required
        />
      </div>

      <div className="space-y-2">
        <label htmlFor="password" className="block text-sm font-medium text-gray-700">
          비밀번호
        </label>
        <input
          id="password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500"
          required
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full px-4 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:bg-gray-400"
      >
        {isLoading ? '로그인 중...' : '로그인'}
      </button>
    </form>
  )
}
```

---

### Level 3: Advanced Topics - 1,500 words

#### Advanced Topic 1: Library Analytics & Webhooks (Enterprise)

**Library Analytics - 변수/컴포넌트 사용 분석**:

```typescript
// 변수 사용 현황 조회
const response = await fetch(
  'https://api.figma.com/v1/analytics/libraries/LIBRARY_FILE_KEY/variable/usages',
  {
    headers: { 'X-Figma-Token': ENTERPRISE_TOKEN }
  }
)

const analytics = await response.json()
// {
//   "variable_usages": [
//     {
//       "variable_id": "Variable:1:2",
//       "variable_name": "color/primary/500",
//       "collection_name": "Semantic Colors",
//       "files": [
//         {
//           "file_id": "FILE_1",
//           "file_name": "Product App",
//           "node_count": 42,  // 42개의 노드에서 사용
//           "last_modified": "2025-11-16T10:30:00Z"
//         },
//         {
//           "file_id": "FILE_2",
//           "file_name": "Marketing Website",
//           "node_count": 18,
//           "last_modified": "2025-11-15T14:20:00Z"
//         }
//       ]
//     }
//   ]
// }

// 컴포넌트 삽입/분리 분석
const componentAnalytics = await fetch(
  'https://api.figma.com/v1/analytics/libraries/LIBRARY_FILE_KEY/component/actions',
  {
    headers: { 'X-Figma-Token': ENTERPRISE_TOKEN }
  }
)

const componentData = await componentAnalytics.json()
// {
//   "component_actions": [
//     {
//       "component_id": "Component:1:2",
//       "component_name": "Button",
//       "actions": [
//         {
//           "action": "COMPONENT_INSTANCE_CREATED",
//           "file_id": "FILE_1",
//           "file_name": "Product App",
//           "count": 120,
//           "last_action_timestamp": "2025-11-16T10:30:00Z"
//         },
//         {
//           "action": "COMPONENT_INSTANCE_DETACHED",
//           "count": 5,
//           "last_action_timestamp": "2025-11-15T09:15:00Z"
//         }
//       ]
//     }
//   ]
// }
```

**Webhooks v2 - 파일 변경 추적**:

```typescript
// Webhook 생성
const webhookResponse = await fetch(
  'https://api.figma.com/v2/webhooks',
  {
    method: 'POST',
    headers: {
      'X-Figma-Token': FIGMA_TOKEN,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      team_id: 'TEAM_ID',
      event_type: 'FILE_UPDATE',
      endpoint: 'https://myapp.com/webhooks/figma',
      passcode: 'secret-passcode-for-verification',
      description: 'Track design system file updates'
    })
  }
)

// Webhook 조회
const listResponse = await fetch(
  'https://api.figma.com/v2/webhooks?team_id=TEAM_ID',
  {
    headers: { 'X-Figma-Token': FIGMA_TOKEN }
  }
)

// Webhook 이벤트 수신 (백엔드에서)
// POST /webhooks/figma
// {
//   "webhook_id": "webhook:1:2",
//   "team_id": "TEAM_ID",
//   "timestamp": 1637094247000,
//   "event_type": "FILE_UPDATE",
//   "event_payload": {
//     "file_id": "ABC123XYZ",
//     "file_key": "abc123xyz",
//     "file_name": "DesignSystem",
//     "created_at": "2025-11-16T10:30:00Z",
//     "updated_at": "2025-11-16T11:45:00Z",
//     "editor_set": [
//       {
//         "id": "USER_1",
//         "email": "designer@company.com",
//         "handle": "designer"
//       }
//     ],
//     "description": "Updated color tokens"
//   }
// }

// Webhook 이벤트 처리 로직
app.post('/webhooks/figma', (req, res) => {
  const { event_type, event_payload } = req.body

  if (event_type === 'FILE_UPDATE') {
    const { file_id, file_name, editor_set } = event_payload

    // CI/CD 트리거
    console.log(`${file_name} updated by ${editor_set[0].handle}`)

    // 자동 배포
    triggerDesignTokensSync(file_id)
    triggerComponentLibraryBuild(file_id)
  }

  res.json({ status: 'received' })
})
```

#### Advanced Topic 2: Plugin API 심화 (플러그인 개발)

**Plugin Manifest 설정**:

```json
{
  "name": "Design System Manager",
  "id": "1234567890",
  "api": "1.0.0",
  "ui": {
    "width": 400,
    "height": 600
  },
  "permissions": [],
  "main": "code.js",
  "editorType": ["design", "dev"],
  "capabilities": ["inspect", "codegen"],
  "codegenLanguages": ["react", "vue", "swift"],
  "codegenPreferences": [
    {
      "name": "Generate React Component",
      "commandId": "generateReact"
    },
    {
      "name": "Generate Design Tokens",
      "commandId": "generateTokens"
    }
  ]
}
```

**Plugin UI + 로직**:

```typescript
// code.js (Plugin 메인 로직)

import { h, render } from 'https://unpkg.com/preact@10'
import htm from 'https://unpkg.com/htm@3'

const html = htm.bind(h)

interface PluginState {
  selectedTokens: string[]
  exportFormat: 'json' | 'css' | 'tailwind' | 'swift'
  isExporting: boolean
}

function App(props: { state: PluginState; onStateChange: (state: PluginState) => void }) {
  const { state, onStateChange } = props

  const handleExport = async () => {
    onStateChange({ ...state, isExporting: true })

    // 선택된 노드에서 Variables 추출
    const selection = figma.currentPage.selection
    if (selection.length === 0) {
      figma.ui.postMessage({ type: 'error', message: 'No selection' })
      return
    }

    const tokens = extractTokens(selection[0])

    // 형식에 따라 변환
    let output: string
    switch (state.exportFormat) {
      case 'json':
        output = JSON.stringify(tokens, null, 2)
        break
      case 'css':
        output = generateCSSVariables(tokens)
        break
      case 'tailwind':
        output = generateTailwindConfig(tokens)
        break
      case 'swift':
        output = generateSwiftConstants(tokens)
        break
    }

    // 파일로 저장
    figma.ui.postMessage({
      type: 'export-complete',
      data: output,
      format: state.exportFormat
    })

    onStateChange({ ...state, isExporting: false })
  }

  return html`
    <div style={{ padding: '16px' }}>
      <h2>Design Tokens Exporter</h2>

      <select
        value=${state.exportFormat}
        onChange=${(e: any) => onStateChange({ ...state, exportFormat: e.target.value })}
      >
        <option value="json">JSON (DTCG)</option>
        <option value="css">CSS Variables</option>
        <option value="tailwind">Tailwind Config</option>
        <option value="swift">Swift Constants</option>
      </select>

      <button onClick=${handleExport} disabled=${state.isExporting}>
        ${state.isExporting ? 'Exporting...' : 'Export'}
      </button>
    </div>
  `
}

// Plugin 초기화
figma.showUI(__html__)
figma.ui.onmessage = async (msg) => {
  if (msg.type === 'close') {
    figma.closePlugin()
  }
}

// 노드에서 토큰 추출
function extractTokens(node: SceneNode): any {
  // ... implementation
}

function generateCSSVariables(tokens: any): string {
  // ... implementation
}

function generateTailwindConfig(tokens: any): string {
  // ... implementation
}

function generateSwiftConstants(tokens: any): string {
  // ... implementation
}
```

**Plugin에서 Selection 이벤트 처리**:

```typescript
// Selection 변경 감지
figma.on('selectionchange', () => {
  const selection = figma.currentPage.selection
  console.log(`Selected ${selection.length} nodes`)

  // 선택된 노드의 정보를 UI에 전송
  figma.ui.postMessage({
    type: 'selection-changed',
    selection: selection.map(node => ({
      id: node.id,
      name: node.name,
      type: node.type,
      boundVariables: 'boundVariables' in node ? node.boundVariables : {}
    }))
  })
})

// Plugin 데이터 저장 (노드에 메타데이터 저장)
const selectedNode = figma.currentPage.selection[0]
if (selectedNode && 'setPluginData' in selectedNode) {
  selectedNode.setPluginData('figma-plugin-data', JSON.stringify({
    componentMapping: 'Button',
    customName: 'PrimaryButton',
    version: '1.0.0'
  }))

  // 나중에 조회
  const pluginData = selectedNode.getPluginData('figma-plugin-data')
}
```

#### Advanced Topic 3: Multi-Mode Variables 관리 (Light/Dark/HighContrast)

**다중 모드 Design System**:

```typescript
// Variables Collection: Semantic Colors
const collection = {
  id: "VariableCollection:1:2",
  name: "Semantic Colors",
  modes: [
    { modeId: "Mode:1:2", name: "Light" },
    { modeId: "Mode:2:3", name: "Dark" },
    { modeId: "Mode:3:4", name: "HighContrast" }
  ],
  defaultModeId: "Mode:1:2"
}

// 각 변수가 모드별로 다른 값 가짐
const variables = [
  {
    id: "Variable:1:2",
    name: "color/text/primary",
    valuesByMode: {
      "Mode:1:2": { r: 0.0, g: 0.0, b: 0.0, a: 1 },      // Light: 검은색
      "Mode:2:3": { r: 1.0, g: 1.0, b: 1.0, a: 1 },      // Dark: 흰색
      "Mode:3:4": { r: 0.0, g: 0.0, b: 0.0, a: 1 }       // HighContrast: 검은색
    }
  },
  {
    id: "Variable:2:3",
    name: "color/background/surface",
    valuesByMode: {
      "Mode:1:2": { r: 1.0, g: 1.0, b: 1.0, a: 1 },      // Light: 흰색
      "Mode:2:3": { r: 0.1, g: 0.1, b: 0.1, a: 1 },      // Dark: 어두운회색
      "Mode:3:4": { r: 0.95, g: 0.95, b: 0.95, a: 1 }    // HighContrast: 밝은회색
    }
  },
  {
    id: "Variable:3:4",
    name: "color/border/default",
    valuesByMode: {
      "Mode:1:2": { r: 0.9, g: 0.9, b: 0.9, a: 1 },      // Light: 밝은회색
      "Mode:2:3": { r: 0.3, g: 0.3, b: 0.3, a: 1 },      // Dark: 어두운회색
      "Mode:3:4": { r: 0.0, g: 0.0, b: 0.0, a: 1 }       // HighContrast: 검은색
    }
  }
]

// CSS 변수 생성 (다중 모드)
function generateCSSVariablesMultiMode(variables: any[]): string {
  let css = ':root {\n'

  // Light 모드 (기본)
  css += '  /* Light Mode (Default) */\n'
  for (const variable of variables) {
    const lightValue = variable.valuesByMode['Mode:1:2']
    css += `  --${variable.name}: rgba(${Math.round(lightValue.r * 255)}, ${Math.round(lightValue.g * 255)}, ${Math.round(lightValue.b * 255)}, ${lightValue.a});\n`
  }

  css += '}\n\n'

  // Dark 모드
  css += '@media (prefers-color-scheme: dark) {\n'
  css += '  :root {\n'
  css += '    /* Dark Mode */\n'
  for (const variable of variables) {
    const darkValue = variable.valuesByMode['Mode:2:3']
    css += `    --${variable.name}: rgba(${Math.round(darkValue.r * 255)}, ${Math.round(darkValue.g * 255)}, ${Math.round(darkValue.b * 255)}, ${darkValue.a});\n`
  }
  css += '  }\n'
  css += '}\n\n'

  // High Contrast 모드
  css += '@media (prefers-contrast: more) {\n'
  css += '  :root {\n'
  css += '    /* High Contrast Mode */\n'
  for (const variable of variables) {
    const hcValue = variable.valuesByMode['Mode:3:4']
    css += `    --${variable.name}: rgba(${Math.round(hcValue.r * 255)}, ${Math.round(hcValue.g * 255)}, ${Math.round(hcValue.b * 255)}, ${hcValue.a});\n`
  }
  css += '  }\n'
  css += '}\n'

  return css
}

// 결과 CSS
/*
:root {
  /* Light Mode (Default) */
  --color/text/primary: rgba(0, 0, 0, 1);
  --color/background/surface: rgba(255, 255, 255, 1);
  --color/border/default: rgba(230, 230, 230, 1);
}

@media (prefers-color-scheme: dark) {
  :root {
    /* Dark Mode */
    --color/text/primary: rgba(255, 255, 255, 1);
    --color/background/surface: rgba(26, 26, 26, 1);
    --color/border/default: rgba(77, 77, 77, 1);
  }
}

@media (prefers-contrast: more) {
  :root {
    /* High Contrast Mode */
    --color/text/primary: rgba(0, 0, 0, 1);
    --color/background/surface: rgba(242, 242, 242, 1);
    --color/border/default: rgba(0, 0, 0, 1);
  }
}
*/
```

**Tailwind Config 생성 (다중 모드)**:

```typescript
function generateTailwindConfigMultiMode(variables: any[]): string {
  const lightVars = {}
  const darkVars = {}

  for (const variable of variables) {
    if (variable.name.startsWith('color/')) {
      const colorName = variable.name.replace('color/', '')

      const lightValue = variable.valuesByMode['Mode:1:2']
      const lightHex = rgbaToHex(lightValue)
      lightVars[colorName] = lightHex

      const darkValue = variable.valuesByMode['Mode:2:3']
      const darkHex = rgbaToHex(darkValue)
      darkVars[colorName] = darkHex
    }
  }

  return `module.exports = {
  theme: {
    colors: {
      // Light mode (default)
      ...{
${Object.entries(lightVars).map(([name, hex]) => `        '${name}': '${hex}',`).join('\n')}
      }
    },
    // Dark mode
    extends: {
      darkMode: 'class',
      dark: {
        colors: {
${Object.entries(darkVars).map(([name, hex]) => `          '${name}': '${hex}',`).join('\n')}
        }
      }
    }
  }
}`
}

function rgbaToHex(rgba: { r: number; g: number; b: number; a: number }): string {
  const r = Math.round(rgba.r * 255).toString(16).padStart(2, '0')
  const g = Math.round(rgba.g * 255).toString(16).padStart(2, '0')
  const b = Math.round(rgba.b * 255).toString(16).padStart(2, '0')
  return `#${r}${g}${b}`
}
```

#### Advanced Topic 4: 성능 최적화 (대규모 파일 처리)

**API 요청 최적화**:

```typescript
// 1. 대용량 파일 메타데이터만 먼저 조회
const metadata = await fetch(
  'https://api.figma.com/v1/files/LARGE_FILE_KEY?geometry=false&plugin_data=false',
  { headers: { 'X-Figma-Token': FIGMA_TOKEN } }
).then(r => r.json())

// geometry와 plugin_data를 false로 설정하면 응답 크기 60% 감소

// 2. 필요한 노드만 조회
const specificNodes = await fetch(
  'https://api.figma.com/v1/files/LARGE_FILE_KEY?ids=123:45,123:46,123:47',
  { headers: { 'X-Figma-Token': FIGMA_TOKEN } }
).then(r => r.json())

// 3. 캐싱 전략
const cache = new Map<string, any>()

async function getCachedFile(fileKey: string) {
  if (cache.has(fileKey)) {
    return cache.get(fileKey)
  }

  const data = await fetch(
    `https://api.figma.com/v1/files/${fileKey}`,
    { headers: { 'X-Figma-Token': FIGMA_TOKEN } }
  ).then(r => r.json())

  cache.set(fileKey, data)

  // 5분 후 캐시 만료
  setTimeout(() => cache.delete(fileKey), 5 * 60 * 1000)

  return data
}

// 4. 배치 요청
const imageUrls = await fetch(
  'https://api.figma.com/v1/images/FILE_KEY?ids=1:2,1:3,1:4,1:5&scale=1&format=png',
  { headers: { 'X-Figma-Token': FIGMA_TOKEN } }
).then(r => r.json())
```

#### Advanced Topic 5: Context7 MCP 통합 (최신 정보 유지)

**자동 API 버전 확인**:

```typescript
import { resolve_library_id, get_library_docs } from '@context7/mcp'

async function getFigmaLatestAPI() {
  // Figma REST API 최신 문서 조회
  const libraryId = await resolve_library_id('Figma REST API')
  // 결과: "/websites/developers_figma"

  const restApiDocs = await get_library_docs(libraryId, { tokens: 5000 })

  // 변수 API 업데이트 확인
  const variablesLibraryId = await resolve_library_id('Figma Variables API')
  // 결과: "/figma/rest-api-spec"

  const variablesDocs = await get_library_docs(variablesLibraryId, {
    topic: 'variables',
    tokens: 3000
  })

  // Code Connect 최신 가이드
  const codeConnectId = await resolve_library_id('Code Connect')
  // 결과: "/figma/code-connect"

  const codeConnectDocs = await get_library_docs(codeConnectId, { tokens: 2000 })

  return {
    restApi: restApiDocs,
    variables: variablesDocs,
    codeConnect: codeConnectDocs
  }
}

// 자동 업데이트 예시 코드 생성
async function generateLatestCodeExamples() {
  const docs = await getFigmaLatestAPI()

  // 문서에서 최신 API 버전 추출
  const latestVersion = docs.restApi.match(/API v(\d+)/)?.[1]

  // 변수 API 최대 한계값 (예: 5,000개 → 10,000개 증가)
  const maxVariables = docs.variables.match(/(\d+) variables per collection/)?.[1]

  // 이러한 정보를 자동으로 코드에 반영
  console.log(`Figma API v${latestVersion} with up to ${maxVariables} variables`)
}
```

**Skill에서 Context7 활용**:

```typescript
// moai-domain-figma Skill에서 자동으로 호출
const Figma_REST_API = resolve_library_id("Figma REST API")
const Figma_Plugin_API = resolve_library_id("Figma Plugin API")
const Figma_Variables_API = resolve_library_id("Figma Variables API")
const Code_Connect_Spec = resolve_library_id("Code Connect")

// 각 라이브러리의 최신 문서를 기반으로 코드 생성
```

---

## 🎯 Success Criteria

✅ **Level 1 완료**: Quick Start로 Figma 생태계 전체 이해
✅ **Level 2 완료**: 5가지 핵심 패턴으로 실무 구현
✅ **Level 3 완료**: Enterprise 기능과 성능 최적화

✅ **REST API 완전 커버**: 7개 엔드포인트 + 인증 + 권한
✅ **Variables API 완전 커버**: 4가지 타입 + Collections + Modes + Code Syntax
✅ **Code Connect 완전 커버**: 4개 프레임워크 (React, Vue, SwiftUI, Compose)
✅ **Dev Mode MCP 완전 커버**: 이미지, Assets, CSS 생성
✅ **Design System 아키텍처**: Atomic Design + Naming Convention
✅ **Context7 통합**: 자동 API 버전 갱신

---

## 📚 Additional Resources

### Figma 공식 문서 (Context7)
- `/websites/developers_figma` - REST API, Plugin API
- `/figma/rest-api-spec` - OpenAPI 스펙, TypeScript 타입
- `/websites/figma_plugin-docs` - Plugin API 상세
- `/figma/code-connect` - Design-to-Code 워크플로우

### 표준 및 명세
- **W3C DTCG 2025.10**: https://designtokens.org (Design Tokens 표준)
- **OpenAPI**: Figma REST API 사양
- **Code Connect Spec**: Design-to-Code 매핑 규칙

---

**Last Updated**: 2025-11-16
**Version**: 1.0.0 Complete
**Status**: Production Ready
**Enterprise**: Yes (Variables API, Library Analytics, Webhooks)