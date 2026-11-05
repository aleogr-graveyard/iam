# WorkOS FGA (ex-Warrant)

| Campo | Valor |
|---|---|
| Categoria | Autorização fina como serviço, acoplada à plataforma de identidade B2B da WorkOS (AuthKit, SSO, Directory Sync, RBAC) |
| Site / repositório | https://workos.com/fga · docs: https://workos.com/docs/fga · legado: https://github.com/warrant-dev/warrant |
| Licença e modelo de negócio | SaaS proprietário. Warrant (motor Zanzibar em Go) segue Apache-2.0 no GitHub, **sem manutenção efetiva** (últimos commits: README em 12/2025, dependências em 01/2025). **Preço do FGA não publicado** na página de preços (que lista AuthKit grátis até 1 M MAU, SSO/Directory Sync US$ 125 por conexão, Audit Logs, Radar); tratar como venda consultiva |
| Stack | API REST da WorkOS; SDKs Node, Python, Go, Ruby, PHP, Java, .NET. Warrant legado: Go com MySQL, Postgres ou SQLite |
| Maturidade | Warrant (YC S21) lançado em 2021/2022; adquirido pela WorkOS em 04/2024; FGA (baseado no Warrant) em early access 07/2024, GA 11/2024; **essa versão foi descontinuada em 15/11/2025**; **novo FGA lançado em 02/2026**, redesenhado como extensão do RBAC. Warrant: ~1,3 mil estrelas |
| Acesso | 2026-11-03 |

## Resumo
- O FGA atual não é Zanzibar genérico: é **RBAC hierárquico por recurso**. Tipos de recurso formam árvore (organização como raiz), papéis são escopados a um tipo e podem incluir permissões de tipos filhos, que descem pela hierarquia.
- Esquema definido **no Dashboard**, não em código ("sem DSL de esquema confusa"); instâncias registradas por API.
- A WorkOS recomenda **não** sincronizar entidades de alta cardinalidade (documentos, mensagens, linhas): autorizar pelo "pai gerenciado mais próximo".
- Duas reescritas de produto em dois anos (Warrant → FGA v1 → FGA v2) forçaram migração dos clientes.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | RBAC por organização e por recurso, herança hierárquica (pai único, profundidade limitada). Sem ABAC, sem relações arbitrárias (ReBAC genérico), sem políticas (PBAC). A v1 tinha políticas e esquema textual, removidos |
| 2. Decisões | ◐ | `check` por API; permissões de organização no **JWT** (checagem sem chamada). Listar usuários de um recurso (diretos/indiretos) e recursos de um usuário sob um pai (só atribuições diretas no pai ou filhos; indireto "ainda não suportado"). Sem explicação |
| 3. Políticas | ○ | Sem linguagem; configuração no Dashboard; sem versionamento, testes, simulação ou implantação gradual documentados |
| 4. Dados | ◐ | Recursos e atribuições por API com `external_id`; eventos de ciclo de vida de papéis/permissões; grupos e mapeamento de grupos do IdP para papéis (SSO/SCIM). Escrita dupla é da aplicação |
| 5. Acesso dinâmico | ○ | Sem expiração, pedidos, aprovações, M de N ou delegação no FGA |
| 6. Estados e transições | ○ | Não |
| 7. Padrões | ○ | AuthZEN não encontrado na doc (llms.txt). Escopos OAuth no AuthKit/MCP Auth, fora do FGA |
| 8. Máquinas e agentes | ◐ | Sujeitos hoje: membros de organização e grupos ("outros tipos em breve"); produtos vizinhos MCP Auth, auth.md, Airlock (Intent-Based Access Control) |
| 9. Desempenho e resiliência | ◐ | Somente SaaS; JWT como cache para permissões de organização. Sem PDP local (o Warrant tinha edge agent) |
| 10. Auditoria | ◐ | Produto Audit Logs separado e pago (streaming US$ 125/mês por SIEM; retenção US$ 99/mês por milhão de eventos); log de decisões do FGA não encontrado |
| 11. Experiência e administração | ◐ | Dashboard para tipos, papéis e permissões; custom roles por organização; Admin Portal para o TI do cliente; sem UI para instâncias de recurso ("não há UI no Dashboard para criar recursos") |

## Diferenciais / exclusivos
- **Simplicidade deliberada**: modelo mental de RBAC estendido à hierarquia, adotável incrementalmente sem migração de dados.
- **Integração nativa com identidade B2B**: papéis vindos de grupos do IdP (SSO/SCIM), custom roles por organização, permissões no token.
- Guia explícito de **o que sincronizar e o que manter na aplicação** (baixa vs. alta cardinalidade), com padrão "pai mais próximo".
- `assignment=direct|indirect` na listagem de quem tem acesso, útil para revisão de acessos/compliance.

## O que a comunidade elogia (com evidência)
- Launch HN do Warrant (2022): pedido por componentes de UI de permissões e boa recepção de edge cache + sync para disponibilidade (comunidade).
- Integração com o restante da WorkOS e adoção incremental (blog/changelog WorkOS, **fornecedor**).
- Discussões independentes sobre o FGA v2 não encontradas (busca web e HN).

## Do que a comunidade se queixa (com evidência)
- Launch HN do Warrant (2022, 115 pontos, 60 comentários): forte rejeição a pôr uma chamada remota no caminho de cada requisição ("se vocês caem, eu caio"), dúvida sobre o que acontece se a startup fechar, preocupações de compliance/GDPR por terceirizar a superfície de ataque.
- Esses medos se materializaram: o produto mudou de dono e de modelo; a v1 foi encerrada em 15/11/2025, e a Oso publicou guia para migrar de WorkOS FGA (concorrente, viés).
- O changelog de "policy helpers" da v1 hoje avisa que "não se aplica mais": funcionalidades removidas na reescrita.
- Limitação reconhecida na doc: listar recursos de um usuário não considera acesso indireto.

## Histórico de segurança
- OSV: nenhuma vulnerabilidade para `github.com/warrant-dev/warrant`; GitHub do Warrant sem advisories.
- **Incidente WorkOS (02/11/2026)**: post "WorkOS reports security breach" no HN; o CEO afirmou que os sistemas de produção não foram afetados e que o incidente ficou restrito a sistemas internos de GTM, com uso de um framework de agente ofensivo autônomo; investigação em curso (comentário do CEO, fonte parte interessada; detalhes não verificados de forma independente).
- **Lições:** (1) separar fortemente sistemas internos (vendas, CRM) dos de produção e dos dados de clientes; (2) comunicação de incidente precisa dizer o que se sabe e o que não se sabe (comentários no HN leram "sem evidência de acesso" como "não sabemos"); (3) ataques automatizados por agentes elevam a cadência esperada de testes ofensivos (premissa 16).

## Arquitetura e decisões notáveis
- Entidades: Subjects (membros, grupos), Resources (árvore, pai único, raiz = organização criada automaticamente), Privileges (papéis e permissões escopados a tipo), Assignments (sujeito + papel + recurso).
- Checagem em dois níveis: JWT para permissões da organização (rápido), API para recurso específico.
- `external_id` único por tipo e organização, permitindo IDs repetidos entre tenants.
- Recomendação de hierarquia rasa (2–4 níveis); o Dashboard valida pai único e profundidade máxima.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy nativa: organização é a raiz de toda hierarquia; papéis de ambiente (globais) e custom roles por organização.
- Extensibilidade baixa: sem linguagem, condições ou funções.
- Configuração no Dashboard + Authorization API (papéis e permissões programáveis desde 02/2026).
- Soberania: somente nuvem WorkOS; sem self-hosted. Warrant OSS (self-hosted) parado.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Linguagem:** nenhuma (configuração declarativa no Dashboard). O Warrant legado tinha esquema textual, tuplas ("warrants"), políticas e linguagem de consulta.
- **Consistência:** "a atribuição vale imediatamente; a próxima checagem a inclui" (doc). Tokens de consistência não documentados; JWT pode ficar defasado até renovar.
- **Listagem:** usuários por recurso (direto/indireto); recursos por usuário sob um pai (só direto).
- **Sincronização (foco):** registrar recursos quando a aplicação os cria; manter alta cardinalidade fora; grupos e papéis do IdP sincronizados pela própria WorkOS (vantagem de quem tem AuthN junto).
- **Explicabilidade, testes, simulação:** não encontrados.
- **Avaliação local:** não (apenas JWT).
- **Desempenho publicado:** não encontrado.
- **AuthZEN:** não encontrado.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N):** **não** no FGA. (O produto "Airlock – Intent-Based Access Control" aparece no menu; não foi analisado nesta sessão por falta de documentação lida; marcar como **?**.)
- **Autorização sensível ao estado / transições declarativas:** **não.**

## Implicações para o projeto
1. **Adotar** a integração AuthN↔AuthZ como vantagem de "família de produtos": grupos do IdP → papéis, permissões amplas no token e finas no PDP (alinha com a boa prática de tokens leves do brief).
2. **Adotar** guia explícito de cardinalidade (o que vira dado do AuthZ, o que fica na aplicação) e o padrão "pai mais próximo" como recomendação de modelagem.
3. **Adotar** o modelo "RBAC hierárquico por recurso" como **degrau inicial** do nosso AuthZ (camada 1, usado pela própria plataforma), por trás do contrato que suportará ReBAC/ABAC depois.
4. **Evitar** reescritas que quebram clientes: contrato estável (AuthZEN) + versionamento de API e caminho de migração automática.
5. **Evitar** esquema apenas no Dashboard: oferecer UI e também artefato versionável (código/arquivo) para revisão e CI.
6. **Oportunidade**: self-hosted completo e PDP local respondem à principal objeção histórica do Warrant ("se vocês caem, eu caio").
7. **Oportunidade**: listagem com acesso indireto, explicabilidade e logs de decisão inclusos (aqui ausentes ou pagos à parte).

## Fontes
- https://workos.com/docs/fga.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/quick-start.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/resource-types.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/roles-and-permissions.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/resources.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/assignments.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/access-checks.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/resource-discovery.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/high-cardinality-entities.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/model-your-app-high-cardinality-data.md — 2026-11-03 — fornecedor
- https://workos.com/docs/fga/standalone-integration.md — 2026-11-03 — fornecedor
- https://workos.com/docs/llms.txt — 2026-11-03 — fornecedor (índice; sem AuthZEN)
- https://workos.com/pricing e https://workos.com/pricing.md — 2026-11-03 — fornecedor (FGA ausente)
- https://workos.com/changelog/fine-grained-authorization — 2026-11-03 — fornecedor (FGA v2, 02/2026)
- https://workos.com/blog/february-2026-updates — 2026-11-03 — fornecedor
- https://workos.com/changelog/fine-grained-authorization-is-now-generally-available — 2026-11-03 — fornecedor (GA 11/2024)
- https://workos.com/changelog/fga-policy-helpers — 2026-11-03 — fornecedor (aviso de obsolescência)
- https://www.osohq.com/post/migrating-workos-fga-to-oso — 2026-11-03 — **concorrente** (data de sunset 15/11/2025)
- https://github.com/warrant-dev/warrant e /commits/main — 2026-11-03 — repositório
- https://github.com/warrant-dev/warrant/security — 2026-11-03 — sem advisories
- https://news.ycombinator.com/item?id=32251305 (Launch HN Warrant, 2022) — 2026-11-03 — comunidade
- https://news.ycombinator.com/item?id=50345712 ("WorkOS reports security breach", 02/11/2026) — 2026-11-03 — comunidade + resposta do CEO (parte interessada)
- https://guptadeepak.com/ciam-compass/changelog/workos-acquires-warrant-2024/ — 2026-11-03 — terceiro (via busca; aquisição 04/2024)
- https://api.osv.dev/v1/query (warrant) — 2026-11-03 — base neutra
