# ADR 0023: AuthZ para máquinas e agentes

- Status: aprovado (2026-11-06)
- Escopo: camada posterior ao marco 1.

## Contexto

O brief 4.29 e 5.2 incluem autorização de serviços, agentes de IA e servidores
MCP. A especificação de autorização do MCP (2026-07-28) trata o servidor MCP
como resource server OAuth 2.1 e proíbe repassar o token recebido. A delegação
"agente em nome de usuário" converge para Token Exchange (RFC 8693) com o claim
`act`. Ainda não há padrão aprovado de identidade específica de agente
(`docs/market-study/standards/agents-mcp.md`).

## Decisão

1. **Contas de serviço e agentes são sujeitos** como qualquer usuário.
2. Agente agindo em nome de alguém usa **delegação**
   ([ADR 0020](0020-authz-dynamic-access.md)) e **token exchange**
   (RFC 8693, claim `act`).
3. **Mapeamento AuthZEN ↔ MCP** conforme a especificação MCP 2026-07-28.
4. **Camada posterior.**

## Alternativas

- **Modelo próprio de identidade de agente:** não há padrão aprovado (grupo
  `agentproto` da IETF ativo, perfis concorrentes); o modelo de sujeito comum
  basta.
- **Agente usando a sessão ou o token do usuário:** a especificação do MCP
  proíbe o repasse de token.

## Consequências

- Os limites da delegação (subconjunto, prazo, revogação) valem para agentes.
- A auditoria registra a cadeia de atores: usuário, agente, cliente e decisão.
- O AuthN precisa de token exchange e client credentials nessa camada
  (brief 5.1, item 6).
- Os perfis AuthZEN de MCP e de token exchange estão em rascunho:
  acompanhar antes de implementar ([ADR 0024](0024-authz-standards.md)).
- Testes de abuso de delegação cobrem agentes.
