# ADR 0012: proibições absolutas e alertas críticos

- Status: aprovado (2026-11-06)

## Contexto

A premissa 4 manda alertar em vez de bloquear. O brief 4.6 propõe uma lista
curta de proibições absolutas para configurações sem nenhum uso legítimo
(pergunta 13). Alertas críticos exigem ciência registrada, e o modo estrito os
transforma em erros (brief 4.6).

## Decisão

**Proibições absolutas** (bloqueadas sempre):

| Proibição |
|---|
| `alg: none` ou algoritmo fora da lista |
| `redirect_uri` com curinga ou em HTTP fora do loopback |
| Fluxos implicit e ROPC |
| Cliente público sem PKCE, ou PKCE `plain` |
| Senha reversível ou hash fora da lista |
| Credencial, segredo ou chave padrão |
| Master e operadores sem MFA resistente a phishing |
| Vinculação automática de contas por e-mail não verificado |
| Autorização fail-open |
| Auditoria administrativa ou de segurança desligada |
| Endpoints de autenticação sem TLS fora de desenvolvimento |

**Alertas críticos** (não são proibições):

- recuperação mais fraca que o MFA;
- métodos fracos como fator único;
- tokens de longa duração;
- login nativo via API.

## Alternativas

- **Só alertas para tudo** (premissa 4 sem exceções): essas configurações não
  têm uso legítimo e cada uma é causa conhecida de incidentes
  (`docs/market-study/summary.md`, seção 4).

## Consequências

- As proibições são verificadas ao carregar a configuração e pela API
  (design, seção 10), em todos os níveis da hierarquia.
- Cada proibição tem teste que tenta configurá-la e verifica a recusa; os
  ataques correspondentes são testes de abuso (design, seção 11).
- Alertas críticos seguem o brief 4.6: ciência registrada na auditoria, com
  expiração, e modo estrito opcional.
