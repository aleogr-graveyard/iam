# ADR 0019: explicabilidade, simulação, testes, versões e governança de políticas

- Status: aprovado (2026-11-06)
- Escopo: no marco 1, versões validadas e explicação básica
  ([ADR 0025](0025-authz-milestone-1-scope.md)); o resto em camadas seguintes.

## Contexto

O brief 4.29 lista como capacidades explicabilidade, simulação e testes antes
de aplicar, versionamento com implantação gradual e governança (quem escreve,
revisa, testa e versiona). O brief 4.5 prevê pisos de segurança herdados. O
estudo de mercado recomenda políticas como código revisável em PR, com replay
de decisões antes de aplicar (`docs/market-study/summary.md`, seção 2).

## Decisão

| Tema | Decisão |
|---|---|
| Explicação | Políticas determinantes + caminho de relações; detalhe só para administradores ou diagnóstico |
| Simulação | "E se?" com uma versão em rascunho |
| Testes | Arquivos de teste de políticas executados por CLI e no CI do cliente |
| Versões | Imutáveis, com ativação e rollback |
| Implantação | Modo sombra antes de ativar |
| Autoria | Políticas como código ou via API e painel |
| Hierarquia | Pisos por hierarquia: `forbid` de nível superior vence |
| Validação | Por schema já; análise formal depois |

## Alternativas

- **Explicação detalhada para qualquer chamador:** exporia a estrutura de
  políticas e relações; o detalhe fica restrito.
- **Análise formal desde o início** (equivalência, subsunção, como no SymCC do
  Cedar): adiada; a validação por schema vem primeiro.

## Consequências

- Toda ativação e todo rollback geram evento de auditoria
  ([ADR 0005](0005-state-tables-and-hash-chained-audit.md)).
- O modo sombra avalia a versão candidata ao lado da ativa e registra as
  divergências.
- Teste de abuso: política de nível inferior não anula `forbid` de nível
  superior.
- O formato dos arquivos de teste faz parte do contrato C10 (design, seção 8).
