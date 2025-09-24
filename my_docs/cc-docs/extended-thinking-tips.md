# 확장 사고 팁

export const TryInConsoleButton = ({userPrompt, systemPrompt, maxTokens, thinkingBudgetTokens, buttonVariant = "primary", children}) => {
const url = new URL("https://console.anthropic.com/workbench/new");
if (userPrompt) {
url.searchParams.set("user", userPrompt);
}
if (systemPrompt) {
url.searchParams.set("system", systemPrompt);
}
if (maxTokens) {
url.searchParams.set("max_tokens", maxTokens);
}
if (thinkingBudgetTokens) {
url.searchParams.set("thinking.budget_tokens", thinkingBudgetTokens);
}
return <a href={url.href} className={`btn size-xs ${buttonVariant}`} style={{
    margin: "-0.25rem -0.5rem"
  }}>
{children || "Try in Console"}{" "}
<Icon icon="arrow-right" color="currentColor" size={14} />
</a>;
};

이 가이드는 Claude의 확장 사고 기능을 최대한 활용하기 위한 고급 전략과 기술을 제공합니다. 확장 사고를 통해 Claude는 복잡한 문제를 단계별로 해결하여 어려운 작업에서 성능을 향상시킬 수 있습니다.

확장 사고를 사용해야 하는 시기를 결정하는 지침은 [확장 사고 모델](/ko/docs/about-claude/models/extended-thinking-models)을 참조하세요.

## 시작하기 전에

이 가이드는 이미 확장 사고 모드를 사용하기로 결정했고, [확장 사고 시작하기](/ko/docs/about-claude/models/extended-thinking-models#getting-started-with-extended-thinking-models)에 관한 기본 단계와 [확장 사고 구현 가이드](/ko/docs/build-with-claude/extended-thinking)를 검토했다고 가정합니다.

### 확장 사고를 위한 기술적 고려사항

- 사고 토큰의 최소 예산은 1024 토큰입니다. 최소 사고 예산으로 시작하여 필요와 작업 복잡성에 따라 점진적으로 증가시키는 것이 좋습니다.
- 최적의 사고 예산이 32K를 초과하는 워크로드의 경우, 네트워킹 문제를 방지하기 위해 [배치 처리](/ko/docs/build-with-claude/batch-processing)를 사용하는 것이 좋습니다. 모델이 32K 토큰 이상 사고하도록 요청하면 장시간 실행되는 요청이 발생하여 시스템 시간 초과 및 열린 연결 제한에 부딪힐 수 있습니다.
- 확장 사고는 영어에서 가장 잘 작동하지만, 최종 출력은 [Claude가 지원하는 모든 언어](/ko/docs/build-with-claude/multilingual-support)로 가능합니다.
- 최소 예산 미만의 사고가 필요한 경우, 표준 모드에서 사고를 끄고 XML 태그(예: `<thinking>`)를 사용한 전통적인 사고 연쇄 프롬프팅을 사용하는 것이 좋습니다. [사고 연쇄 프롬프팅](/ko/docs/build-with-claude/prompt-engineering/chain-of-thought)을 참조하세요.

## 확장 사고를 위한 프롬프팅 기법

### 먼저 일반적인 지침을 사용한 다음 더 단계별 지침으로 문제 해결

Claude는 종종 단계별 처방적 지침보다 작업에 대해 깊이 생각하라는 높은 수준의 지침으로 더 나은 성능을 보입니다. 문제 접근에 대한 모델의 창의성은 인간이 최적의 사고 과정을 처방하는 능력을 초과할 수 있습니다.

예를 들어, 다음과 같이 하는 대신:

<CodeGroup>
  ```text User
  이 수학 문제를 단계별로 생각해보세요: 
  1. 먼저, 변수를 식별하세요
  2. 그런 다음, 방정식을 설정하세요
  3. 다음으로, x에 대해 풀어보세요
  ...
  ```
</CodeGroup>

다음과 같이 고려해보세요:

<CodeGroup>
  ```text User
  이 수학 문제에 대해 철저하고 상세하게 생각해주세요.
  여러 접근 방식을 고려하고 완전한 추론 과정을 보여주세요.
  첫 번째 접근 방식이 작동하지 않으면 다른 방법을 시도해보세요.
  ```

<CodeBlock
filename={
<TryInConsoleButton
userPrompt={
`이 수학 문제에 대해 철저하고 상세하게 생각해주세요.
여러 접근 방식을 고려하고 완전한 추론 과정을 보여주세요.
첫 번째 접근 방식이 작동하지 않으면 다른 방법을 시도해보세요.`
}
thinkingBudgetTokens={16000}

>

    Try in Console

  </TryInConsoleButton>
}
  />
</CodeGroup>

그렇다고 해도, Claude는 필요할 때 복잡한 구조화된 실행 단계를 효과적으로 따를 수 있습니다. 이 모델은 이전 버전보다 더 복잡한 지침이 포함된 더 긴 목록도 처리할 수 있습니다. 더 일반화된 지침으로 시작한 다음, Claude의 사고 출력을 읽고 반복하여 거기서부터 사고를 이끌기 위한 더 구체적인 지침을 제공하는 것이 좋습니다.

### 확장 사고를 통한 멀티샷 프롬프팅

[멀티샷 프롬프팅](/ko/docs/build-with-claude/prompt-engineering/multishot-prompting)은 확장 사고와 잘 작동합니다. Claude에게 문제를 생각하는 방법의 예를 제공하면, 확장 사고 블록 내에서 유사한 추론 패턴을 따를 것입니다.

확장 사고 시나리오에서 `<thinking>` 또는 `<scratchpad>`와 같은 XML 태그를 사용하여 이러한 예제에서 확장 사고의 표준 패턴을 나타내는 방식으로 프롬프트에 몇 가지 예제를 포함할 수 있습니다.

Claude는 이 패턴을 공식적인 확장 사고 과정으로 일반화할 것입니다. 그러나 Claude가 가장 적합하다고 생각하는 방식으로 자유롭게 생각하도록 하면 더 나은 결과를 얻을 수 있습니다.

예시:

<CodeGroup>
  ```text User
  수학 문제를 푸는 방법을 보여드리겠습니다. 그런 다음 비슷한 문제를 풀어주세요.

문제 1: 80의 15%는 얼마인가요?

  <thinking>
  80의 15%를 찾으려면:
  1. 15%를 소수로 변환: 15% = 0.15
  2. 곱하기: 0.15 × 80 = 12
  </thinking>

답은 12입니다.

이제 이 문제를 풀어보세요:
문제 2: 240의 35%는 얼마인가요?

````

<CodeBlock
  filename={
<TryInConsoleButton
  userPrompt={
    `수학 문제를 푸는 방법을 보여드리겠습니다. 그런 다음 비슷한 문제를 풀어주세요.

문제 1: 80의 15%는 얼마인가요?

<thinking>
80의 15%를 찾으려면:
1. 15%를 소수로 변환: 15% = 0.15
2. 곱하기: 0.15 × 80 = 12
</thinking>

답은 12입니다.

이제 이 문제를 풀어보세요:
문제 2: 240의 35%는 얼마인가요?`
  }
  thinkingBudgetTokens={16000}
>
  Try in Console
</TryInConsoleButton>
}
/>
</CodeGroup>

### 확장 사고로 지시 따르기 최대화하기

Claude는 확장 사고가 활성화되었을 때 지시 따르기가 크게 향상됩니다. 모델은 일반적으로:

1. 확장 사고 블록 내에서 지시에 대해 추론
2. 응답에서 해당 지시를 실행

지시 따르기를 최대화하려면:

* 원하는 것에 대해 명확하고 구체적으로 설명하세요
* 복잡한 지시의 경우, Claude가 체계적으로 작업해야 할 번호가 매겨진 단계로 나누는 것을 고려하세요
* Claude가 확장 사고에서 지시를 완전히 처리할 수 있는 충분한 예산을 허용하세요

### 확장 사고를 사용하여 Claude의 행동 디버깅 및 조정하기

Claude의 사고 출력을 사용하여 Claude의 논리를 디버깅할 수 있지만, 이 방법이 항상 완벽하게 신뢰할 수 있는 것은 아닙니다.

이 방법론을 최대한 활용하려면 다음 팁을 권장합니다:

* Claude의 확장 사고를 사용자 텍스트 블록에 다시 전달하는 것은 권장하지 않습니다. 이는 성능을 향상시키지 않고 실제로 결과를 저하시킬 수 있습니다.
* 확장 사고를 미리 채우는 것은 명시적으로 허용되지 않으며, 사고 블록 이후에 나오는 모델의 출력 텍스트를 수동으로 변경하면 모델 혼란으로 인해 결과가 저하될 가능성이 높습니다.

확장 사고가 꺼져 있을 때는 표준 `assistant` 응답 텍스트 [미리 채우기](/ko/docs/build-with-claude/prompt-engineering/prefill-claudes-response)가 여전히 허용됩니다.

<Note>
때로는 Claude가 확장 사고를 어시스턴트 출력 텍스트에서 반복할 수 있습니다. 깔끔한 응답을 원한다면, Claude에게 확장 사고를 반복하지 말고 답변만 출력하도록 지시하세요.
</Note>

### 긴 출력과 장문형 사고의 최대 활용

데이터셋 생성 사용 사례의 경우, "매우 상세한 테이블을 만들어주세요..."와 같은 프롬프트를 사용하여 포괄적인 데이터셋을 생성해보세요.

더 긴 확장 사고 블록과 더 상세한 응답을 생성하고 싶은 상세한 콘텐츠 생성과 같은 사용 사례의 경우, 다음 팁을 시도해보세요:

* 최대 확장 사고 길이를 늘리고 명시적으로 더 긴 출력을 요청하세요
* 매우 긴 출력(20,000+ 단어)의 경우, 단락 수준까지 단어 수가 포함된 상세한 개요를 요청하세요. 그런 다음 Claude에게 단락을 개요에 인덱싱하고 지정된 단어 수를 유지하도록 요청하세요

<Warning>
단순히 토큰 출력을 위해 Claude가 더 많은 토큰을 출력하도록 밀어붙이는 것은 권장하지 않습니다. 대신, 작은 사고 예산으로 시작하여 사용 사례에 맞는 최적의 설정을 찾기 위해 필요에 따라 증가시키는 것이 좋습니다.
</Warning>

다음은 더 긴 확장 사고로 인해 Claude가 뛰어난 성능을 보이는 사용 사례 예시입니다:

<AccordionGroup>
<Accordion title="복잡한 STEM 문제">
  복잡한 STEM 문제는 Claude가 정신적 모델을 구축하고, 전문 지식을 적용하며, 순차적 논리 단계를 거쳐 작업하는 것을 요구합니다. 이는 더 긴 추론 시간의 혜택을 받는 프로세스입니다.

  <Tabs>
    <Tab title="표준 프롬프트">
      <CodeGroup>
        ```text User
        정사각형 안에서 튀어오르는 노란색 공에 대한 파이썬 스크립트를 작성하세요.
        충돌 감지를 적절히 처리해야 합니다.
        정사각형이 천천히 회전하도록 만드세요.
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt={
            `정사각형 안에서 튀어오르는 노란색 공에 대한 파이썬 스크립트를 작성하세요.
충돌 감지를 적절히 처리해야 합니다.
정사각형이 천천히 회전하도록 만드세요.`
          }
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        이 간단한 작업은 일반적으로 몇 초 정도의 사고 시간만 필요합니다.
      </Note>
    </Tab>

    <Tab title="향상된 프롬프트">
      <CodeGroup>
        ```text User
        테서랙트(4차원 초입방체) 안에서 튀어오르는 노란색 공에 대한 파이썬 스크립트를 작성하세요.
        충돌 감지를 적절히 처리해야 합니다.
        테서랙트가 천천히 회전하도록 만드세요.
        공이 테서랙트 안에 머물도록 확인하세요.
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt={
            `테서랙트(4차원 초입방체) 안에서 튀어오르는 노란색 공에 대한 파이썬 스크립트를 작성하세요.
충돌 감지를 적절히 처리해야 합니다.
테서랙트가 천천히 회전하도록 만드세요.
공이 테서랙트 안에 머물도록 확인하세요.`
          }
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        이 복잡한 4D 시각화 과제는 Claude가 수학적 및 프로그래밍 복잡성을 해결하면서 긴 확장 사고 시간을 최대한 활용합니다.
      </Note>
    </Tab>
  </Tabs>
</Accordion>

<Accordion title="제약 최적화 문제">
  제약 최적화는 Claude가 여러 경쟁 요구 사항을 동시에 충족시키도록 도전합니다. 이는 모델이 각 제약 조건을 체계적으로 해결할 수 있도록 긴 확장 사고 시간을 허용할 때 가장 잘 수행됩니다.

  <Tabs>
    <Tab title="표준 프롬프트">
      <CodeGroup>
        ```text User
        일주일 일본 여행 계획을 세워주세요.
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt="일주일 일본 여행 계획을 세워주세요."
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        이 열린 요청은 일반적으로 몇 초 정도의 사고 시간만 필요합니다.
      </Note>
    </Tab>

    <Tab title="향상된 프롬프트">
      <CodeGroup>
        ```text User
        다음 제약 조건을 가진 7일간의 일본 여행 계획을 세워주세요:
        - 예산 $2,500
        - 도쿄와 교토를 반드시 포함해야 함
        - 채식주의 식단을 수용해야 함
        - 쇼핑보다 문화 체험 선호
        - 하루는 반드시 하이킹 포함
        - 장소 간 이동 시간이 하루에 2시간을 넘지 않아야 함
        - 매일 오후에 집에 전화할 수 있는 자유 시간 필요
        - 가능한 한 군중을 피해야 함
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt={
            `다음 제약 조건을 가진 7일간의 일본 여행 계획을 세워주세요:
- 예산 $2,500
- 도쿄와 교토를 반드시 포함해야 함
- 채식주의 식단을 수용해야 함
- 쇼핑보다 문화 체험 선호
- 하루는 반드시 하이킹 포함
- 장소 간 이동 시간이 하루에 2시간을 넘지 않아야 함
- 매일 오후에 집에 전화할 수 있는 자유 시간 필요
- 가능한 한 군중을 피해야 함`
          }
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        여러 제약 조건의 균형을 맞추기 위해, Claude는 모든 요구 사항을 최적으로 충족시키는 방법을 생각할 더 많은 공간이 주어질 때 자연스럽게 가장 좋은 성능을 발휘합니다.
      </Note>
    </Tab>
  </Tabs>
</Accordion>

<Accordion title="사고 프레임워크">
  구조화된 사고 프레임워크는 Claude에게 따라야 할 명시적인 방법론을 제공하며, Claude가 각 단계를 따를 수 있는 긴 확장 사고 공간이 주어질 때 가장 잘 작동할 수 있습니다.

  <Tabs>
    <Tab title="표준 프롬프트">
      <CodeGroup>
        ```text User
        Microsoft가 2027년까지 맞춤형 의학 시장에 진출하기 위한
        포괄적인 전략을 개발하세요.
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt={
            `Microsoft가 2027년까지 맞춤형 의학 시장에 진출하기 위한
포괄적인 전략을 개발하세요.`
          }
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        이 광범위한 전략적 질문은 일반적으로 몇 초 정도의 사고 시간만 필요합니다.
      </Note>
    </Tab>

    <Tab title="향상된 프롬프트">
      <CodeGroup>
        ```text User
        Microsoft가 2027년까지 맞춤형 의학 시장에 진출하기 위한
        포괄적인 전략을 개발하세요.

        다음으로 시작하세요:
        1. 블루 오션 전략 캔버스
        2. 포터의 다섯 가지 힘을 적용하여 경쟁 압력 식별

        다음으로, 규제 및 기술 변수를 기반으로 한 네 가지
        뚜렷한 미래를 가진 시나리오 계획 연습을 수행하세요.

        각 시나리오에 대해:
        - Ansoff 매트릭스를 사용하여 전략적 대응 개발

        마지막으로, 세 가지 지평선 프레임워크를 적용하여:
        - 전환 경로 매핑
        - 각 단계에서 잠재적인 파괴적 혁신 식별
        ```

        <CodeBlock
          filename={
        <TryInConsoleButton
          userPrompt={
            `Microsoft가 2027년까지 맞춤형 의학 시장에 진출하기 위한
포괄적인 전략을 개발하세요.

다음으로 시작하세요:
1. 블루 오션 전략 캔버스
2. 포터의 다섯 가지 힘을 적용하여 경쟁 압력 식별

다음으로, 규제 및 기술 변수를 기반으로 한 네 가지
뚜렷한 미래를 가진 시나리오 계획 연습을 수행하세요.

각 시나리오에 대해:
- Ansoff 매트릭스를 사용하여 전략적 대응 개발

마지막으로, 세 가지 지평선 프레임워크를 적용하여:
- 전환 경로 매핑
- 각 단계에서 잠재적인 파괴적 혁신 식별`
          }
          thinkingBudgetTokens={16000}
        >
          Try in Console
        </TryInConsoleButton>
      }
        />
      </CodeGroup>

      <Note>
        순차적으로 적용해야 하는 여러 분석 프레임워크를 지정함으로써, Claude가 각 프레임워크를 체계적으로 작업함에 따라 사고 시간이 자연스럽게 증가합니다.
      </Note>
    </Tab>
  </Tabs>
</Accordion>
</AccordionGroup>

### 일관성과 오류 처리 향상을 위해 Claude가 자신의 작업을 반성하고 확인하도록 하기

일관성을 향상시키고 오류를 줄이기 위해 간단한 자연어 프롬프팅을 사용할 수 있습니다:

1. Claude에게 작업을 완료하기 전에 간단한 테스트로 작업을 확인하도록 요청하세요
2. 모델에게 이전 단계가 예상된 결과를 달성했는지 분석하도록 지시하세요
3. 코딩 작업의 경우, Claude에게 확장 사고에서 테스트 케이스를 실행하도록 요청하세요

예시:

<CodeGroup>
```text User
숫자의 팩토리얼을 계산하는 함수를 작성하세요.
완료하기 전에 다음 테스트 케이스로 솔루션을 확인해주세요:
- n=0
- n=1
- n=5
- n=10
그리고 발견한 문제를 수정하세요.
````

<CodeBlock
filename={
<TryInConsoleButton
userPrompt={
`숫자의 팩토리얼을 계산하는 함수를 작성하세요.
완료하기 전에 다음 테스트 케이스로 솔루션을 확인해주세요:

- n=0
- n=1
- n=5
- n=10
  그리고 발견한 문제를 수정하세요.`
  }
  thinkingBudgetTokens={16000}
  >
      Try in Console
    </TryInConsoleButton>
  }
    />
  </CodeGroup>
