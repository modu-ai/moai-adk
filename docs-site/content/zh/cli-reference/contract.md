---
title: moai contract 自主执行契约
weight: 100
draft: false
---

`moai contract` 用于验证、显示和签署 SPEC 的自主执行契约（`.moai/specs/<SPEC-ID>/contract.yaml`）。契约写明代理无需人工确认即可执行的操作（actions）、可写入的路径与绝不能触碰的路径（ownership）、需要停下并交还给人处理的条件（escalate_on），以及预算（budget）。签名通过哈希把这些内容与 `acceptance.md` 绑定在一起，签名后只要其中任何一方发生变化，`verify` 就会立即报告不一致。

{{< callout type="info" >}}
在当前版本中，签名**不能替代实现启动审批（Implementation Kickoff Approval）。** 契约是用来记录和验证的手段，人工审批关卡依然保留。
{{< /callout >}}

## 子命令

| 命令 | 说明 | 退出码 |
|------|------|--------|
| `moai contract verify <SPEC-ID>` | 检查契约。只读，可安全地在钩子中调用 | 0 有效 · 1 无效 · 2 用法或 I/O 错误 |
| `moai contract show <SPEC-ID>` | 输出契约各节、签名状态和派生集合（如实际生效的禁写路径） | 0 已输出 · 2 用法或 I/O 错误 |
| `moai contract sign <SPEC-ID>...` | 签署契约 | 0 已签署 · 1 被拒绝 · 2 用法或 I/O 错误 |

`verify` 和 `show` 支持 `--json`，输出机器可读的 JSON 对象。`verify` 只通过封闭的原因代码集合（`unsigned`、`acceptance_hash_mismatch`、`contract_digest_mismatch` 等）报告无效原因。

## moai contract sign

```bash
moai contract sign SPEC-AUTH-001
moai contract sign SPEC-AUTH-001 --resign
moai contract sign SPEC-AUTH-001 --signer llm \
  --receipt .moai/specs/SPEC-AUTH-001/kickoff-receipt.json
```

| 参数 | 说明 |
|------|------|
| `--signer <human\|llm\|llm+jev>` | 签署者。默认走人工（`human`）路径 |
| `--receipt <path>` | 启动回执路径（相对于项目根目录）。`llm` 和 `llm+jev` 签署时必需 |
| `--resign` | 在 `acceptance.md` 变更后重新签署已签名的契约。之前的签名记录为 `supersedes` |

**人工路径。** 仅在交互式终端中运行。它会先显示签署摘要，只有在你输入确认令牌（SPEC ID；一次签署多个时为 `sign N contracts`）后才会签署。如果设置了表示代理运行的环境变量，或标准输入不是终端，则在提示之前直接拒绝。一次签署多个 SPEC 需要开启 `workflow.autonomy.contract.batch_sign`。

**回执路径。** 同时指定 `--signer llm` 或 `--signer llm+jev` 与 `--receipt` 时，无需终端确认，而是以启动回执为依据签署。该路径仅在 `workflow.autonomy.mode` 为 `contract` 时可用，每次只签署一个 SPEC。

**拒绝。** 签署被拒绝时不会修改任何文件，并输出一行 `refused <code> (<SPEC-ID>): <原因>`。代码来自封闭集合（`not_tty`、`confirmation_mismatch`、`already_signed`、`plan_audit_not_passing`、`verify_failed` 等）。待签署的文件在写入前会再次验证，只有确认为有效签名时才会写入，作者的注释和空行都会保留。

## 相关文档

- [config 配置节参考 — workflow.yaml autonomy](/zh/advanced/config-sections/#workflowyaml--autonomy)
- [moai spec 文档管理](/zh/cli-reference/spec)
- [自主性层级 (MOAI_AUTONOMY_TIER)](/zh/advanced/autonomy-tier) — 名称相近，但是互不相关的设置
