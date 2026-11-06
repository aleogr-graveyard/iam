# ADR 0005: estado em tabelas e log de auditoria encadeado, sem event sourcing

- Status: aprovado (2026-11-06)

## Contexto

O brief 4.4 deixou em aberto (pergunta 11) event sourcing nos núcleos, como na
Zitadel, ou estado tradicional com log de auditoria separado. O brief 4.10 e
4.11 pedem um log de auditoria somente de inclusão, com encadeamento por hash.
O estudo de mercado (`docs/market-study/summary.md`, seção 5) mostra a Zitadel
saindo do event sourcing por consistência eventual, projeções travadas e
consultas lentas; a próxima versão dela usa tabelas relacionais como verdade e
eventos só para auditoria.

## Decisão

1. **Estado tradicional em tabelas** como fonte da verdade.
2. **Log de auditoria separado**, somente de inclusão, **encadeado por hash**
   (`shared/audit`, eventos versionados: contrato C4).
3. **Outbox para eventos externos.**
4. Cada transição de máquina de estados gera um evento de auditoria
   (brief 4.4).

## Alternativas

- **Event sourcing nos núcleos:** histórico completo por construção, mas com
  consistência eventual entre eventos e projeções, projeções que travam e
  consultas lentas, problemas que levaram a Zitadel a abandoná-lo.

## Consequências

- Unicidade, transações e RLS ficam no próprio banco
  ([ADR 0001](0001-lab-hosting.md)).
- O histórico vem da auditoria, não do estado; não há reconstrução de estado
  a partir de eventos.
- Outbox e filas por polling, sem `LISTEN/NOTIFY`
  ([ADR 0001](0001-lab-hosting.md)).
- Auditoria administrativa ou de segurança não pode ser desligada
  ([ADR 0012](0012-absolute-prohibitions.md)).
- Testes obrigatórios: alteração ou remoção de um registro encadeado é
  detectada; toda ação administrativa e de segurança gera evento.
