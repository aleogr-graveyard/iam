# Permit.io (inclui OPAL e sincronização de dados)

| Campo | Valor |
|---|---|
| Categoria | Plataforma de autorização (SaaS de controle + PDP no ambiente do cliente); OPAL é a camada aberta de distribuição de políticas e dados |
| Site / repositório | https://www.permit.io · https://docs.permit.io · https://github.com/permitio/opal · PDP: https://github.com/permitio/PDP · docs OPAL: https://docs.opal.ac |
| Licença e modelo de negócio | Plano de controle (UI, API, editor de políticas, Elements, logs) é **proprietário e pago por MAU/tenants**. Abertos (Apache-2.0): OPAL, PDP de contêiner (`permitio/pdp-v2`), `cedar-agent`, SDKs e CLI. A empresa diz não ser open-core, mas oferecer "fallback" aberto para sair da plataforma. Nexus PDP (novo PDP, early access) não aparece como aberto |
| Stack | OPAL e PDP em Python (FastAPI/pydantic) + OPA (Go) ou Cedar Agent (Rust). Pub/sub do OPAL com Postgres, Redis ou Kafka. Nexus PDP: folha NATS + banco embutido para grafo + OPA, sem runtime Python |
| Maturidade | OPAL: ~5,5 mil estrelas, ~300 forks, 42 issues abertas; série 0.x (0.10 trouxe pydantic v2 com quebra de compatibilidade). Empresa: rodada de US$ 6M anunciada em 02/2022 (HN). Citações de uso de OPAL por Tesla e outros (depoimentos no site) |
| Acesso | 2026-11-03 |

## Resumo
- Plataforma "full stack": editor de políticas em UI (RBAC, ABAC, ReBAC) que **gera Rego**, PDP rodando ao lado da aplicação (sidecar), componentes de UI embutíveis (Elements) e logs de decisão.
- O diferencial técnico é o **OPAL**: servidor que observa Git e fontes de dados e empurra atualizações por WebSocket para clientes ao lado do OPA/Cedar.
- GitOps: a política gerada pela UI é escrita num repositório do cliente (um ramo por ambiente); PRs e CI podem revisar e promover.
- Fluxos de **solicitação de acesso e aprovação de operação** prontos (Elements), mas com um revisor, sem quórum M de N.
- Em 2026 lançou o **Nexus PDP** (early access): fluxo de mudanças durável e ordenado, convergente; ainda sem suporte a ABAC.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ● | RBAC, ABAC (condition sets, user/resource sets), ReBAC (tuplas, papéis derivados), PBAC via Rego/Cedar customizado no repositório GitOps |
| 2. Decisões | ◐ | `check`, bulk check, `filter_objects`, `/user-permissions`, `/authorized_users`, endpoints AuthZEN (avaliação, lote, buscas). Explicação via Debug Mode (motivo por modelo e código de negação) |
| 3. Políticas | ● | Editor visual, API, Terraform, GitOps com ramo por ambiente, cópia/merge de ambientes, testes (`opa test`, `permit test generate e2e`), **replay de logs de auditoria** contra PDP novo (`permit test run audit`) |
| 4. Dados | ◐ | Usuários, tenants, instâncias, tuplas via API/SDK; fontes externas via fetchers do OPAL; "proxy facts" para read-your-own-writes. Consistência eventual; sem tokens de consistência |
| 5. Acesso dinâmico | ◐ | Access Request (pedir papel), Operation Approval (aprovar ação), papel por tempo via ABAC. Um revisor por pedido; sem M de N; delegação via Elements de gestão de usuários |
| 6. Estados e transições | ○ | Estado só como atributo em condição ABAC; sem modelo de transições |
| 7. Padrões | ◐ | AuthZEN (evaluation, evaluations, search subject/resource/action, discovery) nos dois PDPs. RAR/UMA não encontrados |
| 8. Máquinas e agentes | ◐ | MCP Gateway, Access Request MCP (aberto), "AI Access Control"; MAU inclui serviços e agentes |
| 9. Desempenho e resiliência | ● | PDP local (sidecar), modo offline com backup em disco, Nexus PDP com dados em disco e cache de memória configurável; benchmarks publicados |
| 10. Auditoria | ◐ | Logs de decisão e operacionais, forwarder, retenção 14 dias (Community) / 21 (Pro); Elements de audit log |
| 11. Experiência e administração | ● | Console, Policy Editor, Elements (gestão de usuários, pedidos, aprovações, logs) embutíveis por iframe; SDKs Node, Go, .NET, Java, Python, Ruby |

## Diferenciais / exclusivos
- **UI que gera código**: produto e suporte editam no editor; engenharia revisa o Rego gerado em PR. Ponte rara entre "low-code" e policy-as-code.
- **OPAL**: distribuição em tempo real de política e dados a PDPs locais, com escopos (multi-repositório / multi-tenant) e fetchers plugáveis.
- **Replay de auditoria** para regressão: compara decisões reais das últimas 6–72 h contra um PDP com a política nova.
- **Elements**: UIs embutíveis para o cliente final (gestão de usuários, pedido de acesso, aprovação de operação, audit log).
- **Nexus PDP**: stream por PDP com ACK, aplicação idempotente por timestamp + ID de transação, reconstrução atômica, relatório de defasagem no health.

## O que a comunidade elogia (com evidência)
- PDP como sidecar resolve latência e disponibilidade (depoimento Signify Health na página de preços — **marketing do fornecedor**).
- Editor de ABAC usável por não técnicos (depoimento Maricopa County, idem).
- OPAL visto como solução madura de replicação para OPA (depoimento de engenheiro da Tesla, idem).
- Em discussão de 2022 no HN sobre autorização como serviço, o modelo "PDP local + OPAL" foi apresentado como resposta ao medo de indisponibilidade (comentário do próprio CEO, viés de fornecedor).
- Não encontrei discussões independentes relevantes no Reddit (busca web sem resultados úteis).

## Do que a comunidade se queixa (com evidência)
- Issues mais comentadas do OPAL: falta de testes E2E (#677, 36 comentários), pedido de OpenFGA como policy store (#661, aberta desde 09/2024), limpeza de symlinks com GitHub fora do ar (#634), vazamento de memória no cliente (#138), vulnerabilidades da imagem base (#513), **polling/configuração redundante de fetchers (#120, aberta desde 08/2021)**.
- Quebra de compatibilidade no OPAL 0.10 (pydantic v2 deixa de converter números em string).
- O container PDP guarda todos os dados como um único JSON no OPA: memória ~20× e sem índice para grafo (admitido na doc do Nexus PDP). Após desconexão, refaz o fetch completo.
- Paridade fragmentada: Cloud PDP não roda ABAC nem Rego customizado; Nexus PDP não roda ABAC e **responde "deny" (não erro)** para políticas ABAC.
- Cobrança por MAU mesmo com PDP local (pergunta frequente da própria página de preços).

## Histórico de segurança
- GitHub Security Advisories do `permitio/opal`: nenhum publicado; **não há SECURITY.md** no repositório (página de segurança do GitHub).
- OSV: nenhuma vulnerabilidade para `opal-server`, `opal-client`, `opal-common`, `permit` (PyPI), `permitio` (npm), `permit-golang`.
- Desenho: o OPAL só ativa verificação de identidade dos clientes em "secure mode" (chaves e master token); em dev é opcional. Aviso explícito para desligar `LOG_DIAGNOSE` (pode expor segredos).
- **Lições:** (1) modo seguro não pode ser opcional: autenticação entre plano de controle e PDP deve ser obrigatória desde o início; (2) política de divulgação e SECURITY.md desde o primeiro release; (3) um PDP que nega silenciosamente por recurso não suportado vira incidente de disponibilidade; preferir erro explícito e alerta (premissa 4).

## Arquitetura e decisões notáveis
- Plano de controle SaaS → OPAL server → WebSocket (saída do cliente, atravessa firewalls) → OPAL client → OPA/Cedar no PDP.
- Container PDP: notificação + fetch na API; Nexus PDP: mensagem carrega o dado (sem fetch), posição durável por PDP, detector de lacunas, snapshot para cold start.
- Consistência: eventual com merge convergente (Nexus). Read-your-own-writes só via "proxy facts" no container PDP (o PDP encaminha a escrita e espera receber a atualização, com timeout e política de timeout).
- Offline: PDP grava config (criptografada com chave derivada da API key) e policy store em disco a cada minuto; pode iniciar sem plano de controle.
- Benchmarks publicados (Nexus PDP 0.7.0-beta.50, 2 vCPU): RBAC P50 4,27 ms / P99 5,59 ms (lado servidor); ReBAC profundidade 3 P99 11,28 ms; capacidade 658 req/s RBAC por pod; 3 pods 1.606 req/s. Fonte do fornecedor, metodologia detalhada.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Tenant é objeto de primeira classe; papel por tenant; check informa tenant no recurso. **Esquema de papéis por tenant não suportado** ("tenants são dados, não política"): usar projetos/ambientes separados.
- Extensibilidade: Rego customizado na pasta `custom` do repositório GitOps; fetchers OPAL próprios; Cedar como motor alternativo.
- Soberania: decisões e dados ficam no PDP do cliente, mas a fonte da verdade (API, UI, logs) é SaaS; on-prem e multi-cloud só no Enterprise.
- Saída da plataforma documentada ("OSS fallback"): exportar Rego via GitOps, rodar OPAL server próprio, montar PDP.

## Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Linguagem:** sem linguagem própria; UI/API geram Rego (ou Cedar). Customização direta em Rego.
- **Consistência:** eventual; sem zookies. Proxy facts como mitigação pontual.
- **Listagem:** `/user-permissions`, `/authorized_users`, `filter_objects` (bulk check sobre lista fornecida); buscas AuthZEN. Não gera filtro SQL.
- **Sincronização de dados (foco):** três caminhos: (a) API/SDK de usuários/papéis/tuplas (escrita dupla na aplicação); (b) OPAL data update triggers: serviço publica evento com URL e tópico; cada cliente busca a fonte (fetch provider) e grava no OPA; (c) dados de contexto por requisição. A issue #120 mostra o custo de polling redundante.
- **Explicabilidade:** Debug Mode por PDP, por API ou por requisição; códigos de negação com páginas de ajuda (`no_matching_rules`, `user_not_synced`, `no_role_in_tenant`...). Aumenta latência.
- **Testes:** unitário (`opa test`), integração (`permit test generate e2e`, ambientes copiados), regressão (replay de auditoria).
- **Avaliação local:** sidecar obrigatório para ABAC e Rego customizado; Cloud PDP para casos simples.
- **AuthZEN:** sim, inclusive buscas e discovery.

## Verificação de diferenciais candidatos do projeto
- **Aprovação multipartes (M de N):** **parcial/não.** Há Operation Approval + Approval Management: o pedido vira instância de recurso; o app atribui o papel `_Reviewer_`; ao aprovar, o solicitante ganha `_Approved_`. **Um revisor por pedido** ("Each request needs its own resource instance key and its own reviewer"), sem quórum, sem vínculo criptográfico à ação, sem janela de tempo nativa.
- **Autorização sensível ao estado / transições declarativas:** **não.** Estado só como atributo em condição ABAC; nenhum conceito de estados/transições.

## Implicações para o projeto
1. **Adotar**: UI de políticas que gera artefato versionado e revisável (Git/PR) por ambiente; une premissas 18 e 20 sem abrir mão de policy-as-code.
2. **Adotar**: replay de decisões reais contra política candidata como teste de regressão e "implantação gradual" (domínio 5.2.3).
3. **Adotar** o desenho de sincronização do Nexus PDP (stream durável com ACK por PDP, idempotência por versão, rebuild atômico, health com defasagem) em vez de "notifica e busca".
4. **Evitar** paridade fragmentada entre modos de execução; se um modo não suporta um recurso, recusar a publicação da política ou alertar, nunca negar silenciosamente.
5. **Oportunidade**: aprovação de operação com M de N, vínculo à ação exata e evidência criptográfica: o concorrente mais próximo tem só um revisor.
6. **Oportunidade**: preço por MAU com PDP local é queixa; considerar modelo sem cobrança por usuário na edição self-hosted.
7. **Evitar** autenticação opcional entre plano de controle e PDP.

## Fontes
- https://www.permit.io/pricing — 2026-11-03 — fornecedor (tabela de planos extraída do HTML; preços Startup/Pro não presentes no HTML estático)
- https://www.permit.io/blog/permit-new-pricing-model — 2026-11-03 — fornecedor ("até 10.000 usuários a partir de US$ 150/mês", nov/2024)
- https://docs.permit.io/integrations/gitops/overview/ — 2026-11-03 — fornecedor
- https://docs.permit.io/concepts/pdp/nexus-pdp/ — 2026-11-03 — fornecedor
- https://docs.permit.io/concepts/pdp/nexus-pdp-how-it-works/ — 2026-11-03 — fornecedor
- https://docs.permit.io/concepts/pdp/nexus-pdp-benchmarks/ — 2026-11-03 — fornecedor (benchmark próprio)
- https://docs.permit.io/concepts/pdp/nexus-pdp-feature-parity/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/deploy/offline-mode/ — 2026-11-03 — fornecedor
- https://docs.permit.io/concepts/oss-fallback/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/enforce-permissions/data-filtering/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/manage-data/local-facts-uploader/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/SDLC/authz-testing/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/permit-cli/permit-cli-test/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/use-audit-logs/audit-log-replay/ — 2026-11-03 — fornecedor
- https://docs.permit.io/how-to/use-audit-logs/debug-mode/ — 2026-11-03 — fornecedor
- https://docs.permit.io/concepts/multi-tenant-authorization/ — 2026-11-03 — fornecedor
- https://docs.permit.io/manage-your-account/creating-environments/ — 2026-11-03 — fornecedor
- https://docs.permit.io/overview/access-requests-and-approvals/ — 2026-11-03 — fornecedor
- https://docs.permit.io/embeddable-uis/element/operation-approval/ — 2026-11-03 — fornecedor
- https://docs.permit.io/embeddable-uis/element/approval-management/ — 2026-11-03 — fornecedor
- https://github.com/permitio/opal — 2026-11-03 — repositório
- https://github.com/permitio/opal/issues?q=is%3Aissue+sort%3Acomments-desc — 2026-11-03 — comunidade
- https://github.com/permitio/opal/security — 2026-11-03 — sem advisories, sem SECURITY.md
- https://docs.opal.ac/overview/architecture — 2026-11-03 — fornecedor
- https://docs.opal.ac/overview/scopes — 2026-11-03 — fornecedor
- https://docs.opal.ac/tutorials/trigger_data_updates — 2026-11-03 — fornecedor
- https://docs.opal.ac/getting-started/running-opal/run-opal-server/security-parameters — 2026-11-03 — fornecedor
- https://api.osv.dev/v1/query (pacotes opal-*, permit, permitio, permit-golang) — 2026-11-03 — base neutra
- https://news.ycombinator.com/item?id=34621411 ("Thoughts on Permit.io?") — 2026-11-03 — comunidade (inclui concorrente Cerbos)
- https://news.ycombinator.com/item?id=32251305 (Launch HN Warrant; comentário do CEO da Permit) — 2026-11-03 — comunidade/fornecedor
