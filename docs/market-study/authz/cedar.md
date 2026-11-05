# Cedar

| Campo | Valor |
|---|---|
| Categoria | Linguagem de políticas + motor de avaliação embutível (RBAC/ABAC/ReBAC por hierarquia), sem estado |
| Site / repositório | https://www.cedarpolicy.com — https://github.com/cedar-policy/cedar (Rust) — https://github.com/cedar-policy/cedar-go (Go) |
| Licença e modelo de negócio | Apache 2.0. Criado pela AWS; **CNCF Sandbox desde 08/10/2025**. Monetização indireta pela AWS (Amazon Verified Permissions, Bedrock AgentCore Policy) |
| Stack | Rust (implementação de referência, crates `cedar-policy`, `-core`, `-cli`, `-symcc`, `-formatter`, `cedar-wasm`, language server); Go (`cedar-go`); modelo formal em Lean |
| Maturidade | Aberto em mai/2023; Rust ~1,8k estrelas, `cedar-policy` 4.14.1 (02/11/2026); `cedar-go` ~250 estrelas, v1.9.0 (29/10/2026), 25 importadores no pkg.go.dev. Adotantes citados: AWS, Cloudflare, MongoDB, StrongDM, Cloudinary, Janssen (LF) |
| Acesso | 2026-11-03 |

## Resumo (até 5 linhas)
Linguagem `permit`/`forbid` com escopo `principal, action, resource` + condições `when`/`unless`, avaliada sobre um conjunto de entidades enviado pelo chamador. Desenhada para ser analisável (codificação lógica sound e completa), com especificação em Lean, provas de propriedades e testes diferenciais contra a implementação Rust. Muito rápida (µs). O motor não guarda dados nem relações: a aplicação precisa enviar as entidades relevantes. A versão Go é parcial (sem avaliação parcial, sem templates, validador experimental).

## Funcionalidades por domínio (5.2)
| Domínio | Suporte | Notas |
|---|---|---|
| 1. Modelos | ● | RBAC (grupos via `in`), ABAC (atributos, `context`), ReBAC limitado a hierarquia de entidades e atributos que referenciam entidades; *entity tags* |
| 2. Decisões | ◐ | Check individual; lote e listagem via **TPE** (experimental, Rust) ou `x/exp/batch` (Go); explicabilidade: resposta traz políticas determinantes e erros |
| 3. Políticas | ● | Linguagem própria + formato JSON, schema, validação estrita e por nível, templates (Rust), formatter, linter de idioma via language server; versionamento fica com o usuário (AVP oferece *policy stores*) |
| 4. Dados | ◐ | Sem armazenamento; entidades enviadas por requisição; *entity slicing* (manifesto, agora substituído por TPE) para reduzir o que se envia |
| 5. Acesso dinâmico | ○ | Nada de primeira classe; JIT expressável com `context` + `datetime`/`duration` (RFC 80) |
| 6. Estados e transições | ◐ | Estado como atributo do recurso; sem modelo de transições |
| 7. Padrões | ◐ | AuthZEN não nativo; Amazon Verified Permissions passa no interop via proxy; Janssen Cedarling expõe endpoints AuthZEN |
| 8. Máquinas e agentes | ◐ | Base do Bedrock AgentCore Policy (agentes); autorizador Kubernetes; projetos da comunidade para agentes de código |
| 9. Desempenho e resiliência | ● | Embutido, sem estado, latência de µs; Wasm |
| 10. Auditoria | ○ | Não há; fica com a aplicação (AVP integra CloudTrail) |
| 11. Experiência e administração | ◐ | Playground/tutorial, language server, VS Code; sem UI de administração no OSS |

## Diferenciais / exclusivos
- **Verificação formal:** modelo em Lean com provas de propriedades (ex.: `forbid` vence `permit`, padrão negar), validador provado *sound*; implementação Rust testada por **differential random testing** contra a especificação Lean [paper; blog AWS].
- **Análise de políticas (SymCC):** `cedar-policy-symcc` compila políticas+schema para SMT (cvc5) e verifica: nunca gera erro, sempre permite, sempre nega, **subsunção (implies)**, **equivalência**, **disjunção**, com contraexemplo sintetizado (requisição + entidades) [docs.rs symcc 0.8.1]. Exposto no CLI "experimental" (`cedar symcc ...`).
- **Validação por nível** (*level-based validation*): garante que políticas só leem até N saltos de entidades, permitindo saber o que carregar.
- Semântica deliberadamente restrita (sem laços, sem regex, sem ponto flutuante) para manter decidibilidade.

## O que a comunidade elogia (com evidência)
- Rigor formal e "verification-guided development" (HN #35890784, 99 pontos, 2023).
- Equilíbrio expressividade × analisabilidade (citação de Lucas Käldström no anúncio CNCF — viés: AWS).
- Desempenho e legibilidade frente a Rego/OpenFGA (paper — viés: autores).

## Do que a comunidade se queixa (com evidência)
- Sensação de linguagem "desajeitada" comparada a OPA em casos genéricos (HN #35890784).
- Responsabilidade de montar e enviar entidades a cada requisição; custo em dados grandes reconhecido pela própria doc, mitigado por *slicing*/TPE [docs Cedar; livro Manning].
- Ergonomia de IDs e operadores: issues abertas mais votadas — #64 (IDs mais ergonômicos), #1920 (`anyLike`/`allLike`), #1274 (operador IN em resource), #880 (precisão da avaliação parcial), #2522 (como informar ao usuário o que falta).
- `cedar-go` atrás do Rust: sem avaliação parcial, sem templates, sem CLI/formatter, validador experimental [README cedar-go].
- Limitações de linguagem: sem regex, sem ponto flutuante [docs Bedrock AgentCore].

## Histórico de segurança
| ID | Ano | Classe | Lição para nós |
|---|---|---|---|
| GHSA-c33x-vcv4-vg3g | 2026 | Templates em JSON sem validação de slot-id: *panic* no link e **divergência de decisão** (Baixa) | Formatos alternativos (JSON × texto) precisam da mesma validação; testes diferenciais entre representações |
| GHSA-m84h-6f26-p3p7 | 2026 | Exaustão de memória ao analisar `has` com caminho de atributo aninhado (Alta) | *Fuzzing* de parser e limites de tamanho/profundidade em políticas vindas de tenants |
- `cedar-go`: nenhum advisory encontrado (GitHub do repositório e OSV.dev).

## Arquitetura e decisões notáveis
- Autorização = função pura `(políticas, entidades, requisição) → decisão`; nenhum I/O no motor.
- `forbid` sempre vence; padrão é negar; erros em uma política a fazem ser ignorada (não derrubam a decisão).
- Escopo `principal/action/resource` indexável → avaliação rápida e *policy slicing* (10x–18x mais rápido em média no paper).
- Duas implementações (Rust de produção, Lean de especificação) + DRT; Go é port independente com suíte de corpus noturna.
- Governança migrando para CNCF; próximos passos declarados: Incubation e Graduated.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: a cargo do integrador (um *policy set* por tenant; AVP usa *policy store* por tenant).
- Extensibilidade: tipos de extensão fixos (ipaddr, decimal, datetime); não há funções definidas pelo usuário — escolha deliberada.
- Bancos: nenhum; soberania total quando embutido.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Embutibilidade em Go (`cedar-go`):** v1.x estável, **3 dependências diretas** (go-json-experiment, go-cmp, x/exp) [proxy.golang.org v1.9.0]. Suporta autorizador, JSON, tipos core/extensões, schema (parse/construção); validação em `x/exp/schema` e lote em `x/exp/batch` (fora do semver). Não tem avaliação parcial nem templates.
- **Linguagem:** Cedar + JSON; schema em formato Cedar ou JSON.
- **Consistência:** não se aplica ao motor (sem estado); consistência depende de quem fornece as entidades.
- **Listagem:** Rust `tpe` (*type-aware partial evaluation*, experimental) com `ResourceQueryRequest`, `PrincipalQueryRequest`, `ActionQueryRequest` → resíduo de política; **não gera SQL**, a tradução fica com o integrador. `partial-eval` antigo desaconselhado; `entity-manifest` depreciado em favor de TPE [docs.rs cedar-policy 4.14.1].
- **Explicabilidade:** `Response` com `reason` (IDs das políticas determinantes) e erros; sem trace de avaliação.
- **Testes:** `cedar run-tests` com arquivo JSON de casos (CLI); validação contra schema; pipelines de exemplo da AWS.
- **Análise estática:** validador tipado (strict/permissive/partial), validação por nível, SymCC (SMT), *lint* via language server.
- **Desempenho publicado:** mediana ~4–5 µs (gdrive, crescendo pouco com o número de entidades); **28,7x–35,2x mais rápido que OpenFGA e 42,8x–80,8x mais rápido que Rego** (EC2 m5.4xlarge, Cedar 3.0.1, Rego 0.61.0, só `is_authorized`, sem parsing) [paper OOPSLA 2024 — viés: autores/AWS].
- **AuthZEN:** sem endpoint no OSS. AVP passou 43/43 no cenário todo‑1.1 do interop (via proxy); Cedarling (Janssen) expõe `/access/v1/evaluation(s)`.

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N)? **Não.** Pode-se passar aprovações no `context` e contar com `.containsAll`/tamanho de conjunto, mas sem coleta, vinculação à ação ou verificação criptográfica.
- Autorização sensível ao estado / transições declarativas? **Parcial.** `resource.status == "draft"` é idiomático; ações podem ser nomeadas como transições (`Action::"sign"`), mas não há declaração de estados/transições nem validação de que a transição existe a partir do estado S.

## Implicações para o projeto
1. **Adotar como referência** de design: semântica pura, `forbid` vence, padrão negar, sem laços — e escopo indexável para desempenho previsível.
2. **Oportunidade forte:** `cedar-go` tem 3 dependências e API estável → candidato natural a motor embutido atrás do adaptador (compatível com premissas 22 e 24). Lacunas a cobrir: avaliação parcial/listagem e validação estável em Go.
3. **Adotar** análise de mudanças de política (equivalência, subsunção, "sempre permite") como gate de CI/aprovação de políticas; é diferencial raro.
4. **Adotar** testes diferenciais entre representações (texto × JSON) e entre implementações — os dois advisories de 2026 são de divergência/robustez de parser.
5. **Evitar** depender do motor para buscar dados: definir no nosso produto o carregamento de entidades (equivalente a *slicing*/validação por nível) e a consistência.
6. **Oportunidade:** listagem com saída SQL não existe no ecossistema Cedar — gerar filtros a partir do resíduo do TPE seria diferencial.

## Fontes
- https://github.com/cedar-policy/cedar — 2026-11-03
- https://github.com/cedar-policy/cedar-go e https://pkg.go.dev/github.com/cedar-policy/cedar-go (README, versões) — 2026-11-03
- https://proxy.golang.org/github.com/cedar-policy/cedar-go/@v/v1.9.0.mod — 2026-11-03
- https://docs.rs/cedar-policy/latest/cedar_policy/ (feature flags, TPE) — 2026-11-03
- https://docs.rs/cedar-policy-symcc/latest/cedar_policy_symcc/ — 2026-11-03
- https://docs.rs/crate/cedar-policy-cli/latest — 2026-11-03
- https://docs.rs/crate/cedar-policy-cli/latest/source/src/command/run_test.rs — 2026-11-03 (visto via busca)
- https://arxiv.org/abs/2403.04651 (PDF, seções de avaliação) — 2026-11-03 — viés: autores (AWS)
- https://aws.amazon.com/blogs/opensource/cedar-joins-cncf-as-a-sandbox-project — 2026-11-03 — viés: AWS
- https://www.cncf.io/projects/cedar/ — 2026-11-03 (data de aceite via busca)
- https://infoq.com/news/2026/01/cedar-joins-cncf-sandbox/ — 2026-11-03 — imprensa (visto via busca)
- https://github.com/cedar-policy/cedar/security/advisories — 2026-11-03
- https://osv.dev (crates.io `cedar-policy`; Go `cedar-go`) — 2026-11-03
- https://github.com/cedar-policy/cedar/issues?q=is%3Aissue+is%3Aopen+sort%3Areactions-%2B1-desc — 2026-11-03
- https://docs.cedarpolicy.com/ e /policies/level-validation.html — 2026-11-03
- https://news.ycombinator.com/item?id=35890784 — 2026-11-03 — comunidade
- https://livebook.manning.com/learn/authorization-in-action/chapter-9 — 2026-11-03 (visto via busca)
- https://docs.aws.amazon.com/bedrock-agentcore/latest/devguide/policy-limitations-section.md — 2026-11-03 (visto via busca)
- https://authzen-interop.net/docs/scenarios/todo-1.1/results/avp/ — 2026-11-03
- https://docs.jans.io/head/cedarling/integrations/cedarling-opa/ — 2026-11-03 (visto via busca)
