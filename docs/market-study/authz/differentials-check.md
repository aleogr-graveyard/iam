# Verificação de diferenciais candidatos (AuthZ)

Acesso às fontes: 2026-11-03 a 2026-11-05. Base: documentos deste diretório (`plainid.md`, `axiomatics.md`, `idp-embedded-authz.md`, `approvals-pam.md`, `openfga.md`) e verificações adicionais abaixo.
Limites: a cota de buscas web da sessão esgotou no fim da pesquisa. Cerbos, Oso, Permit.io, Indykite, Styra e SGNL foram verificados por documentação ou imprensa, mas com menos profundidade (ver "Não verificado").

## Pergunta (a)
**Aprovação multipartes M de N, com assinatura WebAuthn por aprovador vinculada à ação exata, é diferencial?**

### Conclusão
**Parcialmente.** O **M de N sozinho não é diferencial**: é recurso de primeira classe em vários produtos de PAM e nuvem. A **combinação** abaixo não foi encontrada em nenhum produto de identidade ou de autorização disponível:
1. serviço genérico, para ações de **aplicações de terceiros**;
2. representação canônica da ação;
3. **assinatura WebAuthn de cada aprovador sobre o hash da ação**;
4. quórum composto por papéis;
5. evidência verificável pelo próprio executor;
6. integração ao contrato AuthZEN.

Há dois precedentes que reduzem a novidade:
- **Turnkey** (infraestrutura de carteiras cripto) já entrega quórum com passkey sobre o corpo exato da requisição.
- **ScrambleID Lockstep** descreve exatamente esse modelo para identidade, mas está **em desenvolvimento** e ainda não é vendido (junho de 2026).

A janela existe, mas é estreita. O diferencial precisa ser a combinação, aberta e auditável, e não o "M de N".

### Evidência: quem tem M de N de primeira classe

| Produto | M de N? | Composição por papéis | Forte autenticação por aprovação | Vínculo à ação exata | Fonte |
|---|---|---|---|---|---|
| Teleport access requests | **Sim** (`thresholds: approve: N, deny: N`) | Sim (`filter` por `reviewer.roles/traits`) | WebAuthn **por requisição** se "MFA for admin actions" estiver ativo; aprovar pedido conta como ação administrativa | Não há evidência de que o desafio contenha o conteúdo da revisão (RFD 0131 diz só "a MFA vale para uma requisição de API") | goteleport.com docs; RFD 0131 |
| HashiCorp Vault control groups (Enterprise) | **Sim** (`approvals = N` por fator; fatores em AND) | Sim (grupo de identidade por fator) | Não encontrado | **Sim, por estado no servidor**: a requisição original fica guardada (wrapping token) e é executada após as autorizações | developer.hashicorp.com |
| GCP Privileged Access Manager | **Sim** (até 5 por nível, 2 níveis; SCC Premium/Enterprise, preview) | Por nível | Não encontrado | Grant de papéis | cloud.google.com |
| AWS Multi-party approval | **Sim** (M de N, mínimo 2 de 3) | Não | Login IAM Identity Center | Operação protegida específica, executada automaticamente no quórum | docs.aws.amazon.com/mpa |
| CyberArk dual control | **Sim** (número de confirmadores na Master Policy) | ? | ? | Acesso a conta | kb.wisc.edu (cliente) |
| GitLab deployment approvals | **Sim** (`required_approvals` por regra) | Sim | — | Job de deploy | docs.gitlab.com |
| Fireblocks | **Sim** (threshold por grupo, AND/OR) | Sim | App móvel | Transação | developers.fireblocks.com |
| **Turnkey** | **Sim** (`consensus`: `approvers.filter(...).count() >= 2`) | Sim (tags) | **WebAuthn/passkey** por requisição | **Sim, criptográfico**: "desafio WebAuthn = SHA-256 do corpo POST"; aprovação por `fingerprint` da atividade; avaliação em enclave com "verifiable policy decisions" | docs.turnkey.com |
| Google Workspace MPA | 2 pessoas | Privilégio de revisor | — | Mudança de configuração | knowledge.workspace.google.com |
| Intune Multi Admin Approval | 2 pessoas | Grupo aprovador | — | Mudança específica | learn.microsoft.com |
| Azure Backup MUA | 2 papéis (via Resource Guard + PIM) | Sim | — | Operação crítica | docs.azure.cn |
| ScrambleID Lockstep | "dois ou mais" | Conjunto, papel ou grupo de aprovadores | "provas criptográficas de dispositivos vinculados (WebAuthn ou chaves de dispositivo)" | "pedido explicitamente escopado", TTL rígido (ex.: 30 min) | scrambleid.com, **status: em desenvolvimento** |

### Evidência: quem NÃO tem (identidade e AuthZ)

| Produto | Situação | Fonte |
|---|---|---|
| Microsoft Entra PIM | **1 de N**: "Requests are resolved by the first approver who approves or denies"; "Multi-step approvals aren't currently supported"; prazo fixo de 24 h | learn.microsoft.com |
| Entra entitlement management | Até 3 estágios sequenciais, "only one of the selected approvers from each stage needs to approve" | docs.azure.cn |
| Keycloak Authorization Services | Aprovação UMA pelo dono do recurso (1 pessoa). A estratégia "Consensus" combina políticas; não é quórum humano | keycloak.org |
| Zitadel, Auth0 (core) | Sem fluxo de aprovação. Auth0 mostra RAR no consentimento e no Guardian, mas para o próprio usuário | docs oficiais |
| PlainID | Nenhuma página de aprovação no índice da documentação (379 entradas). A revisão de políticas é delegada a "GitHub pull requests" | docs.plainid.io |
| Axiomatics | Nada no ADS, no Hub ou no APD. Só obrigações XACML, que o PEP teria de implementar | docs.axiomatics.com |
| OpenFGA | Sem solicitações e aprovações | `openfga.md` |
| Permit.io | Tem "Operation Approval" para uma ação (ex.: transferência alta), mas a documentação não descreve vários aprovadores nem quórum | docs.permit.io |
| ConductorOne | Etapas sequenciais; em revisor do tipo grupo, "only one member needs to complete the task" | c1.ai/docs |
| Opal | Até 3 estágios, com All ou Any por estágio (AND de revisores nomeados, não M de N) | docs.opal.dev |
| Lumos, Okta OIG | Até 3 níveis ou sequências de etapas | support.lumos.com, help.okta.com |

### Padrão (AuthZEN)
O rascunho **AuthZEN Access Request and Approval Profile, Draft 1 (29/10/2026)**:
- Padroniza o pedido após uma "negação requisitável", a tarefa assíncrona (com `current_step`, `total_steps` e `awaiting`) e a **reavaliação** com `context.approval`.
- **Exige** checar a elegibilidade do aprovador: autoaprovação, separação de funções e conflito de interesse.
- **Exige** vincular a aprovação a sujeito, recurso, ação, contexto e expiração ("approval is not a bearer grant").
- Declara **fora de escopo** "workflow engines, approval policy languages". Não fala de quórum, WebAuthn nem assinaturas de aprovadores.

Consequência: o padrão abre espaço para o nosso M de N por dentro do Access Request Service, sem conflito.

### Limitação técnica que afeta o argumento de venda
O WebAuthn não garante que **o autenticador mostrou** a ação ao aprovador:
- As extensões `txAuthSimple`/`txAuthGeneric` nunca foram implementadas pelos navegadores (há uma proposta de revisão de 2024, w3c/webauthn#2022).
- Secure Payment Confirmation cobre só pagamentos.

A assinatura prova que "esta credencial, com verificação de usuário, assinou este hash". **O que a pessoa viu depende da nossa UI.** O marketing não pode prometer "what you see is what you sign" no nível do autenticador.

### Recomendação
Tratar como diferencial **a combinação** e **posicionar contra PIM e IGA**, não contra PAM. Os elementos:
- M de N com composição por papéis e regra de negação;
- assinatura WebAuthn por aprovador sobre a representação canônica;
- execução só no quórum ("segurar a operação", como Vault e AWS MPA);
- pacote de evidências verificável pelo executor, sem confiar na plataforma;
- encaixe no AuthZEN AARP;
- open source.

Antes do lançamento, reavaliar Teleport, ScrambleID Lockstep e Turnkey: são os mais próximos.

## Pergunta (b)
**Autorização declarativa de estados e transições de recursos é diferencial?**

### Conclusão
**Sim, como produto de AuthZ.** Nenhum motor ou plataforma de autorização pesquisado tem **estados e transições como conceito declarativo de primeira classe**. Isso inclui OpenFGA, Cerbos, Axiomatics, PlainID, Keycloak AS, Permit.io, Oso, Zitadel, Auth0 e Entra. Todos tratam estado como **atributo** do recurso (ex.: `request.resource.attr.status == "PENDING_APPROVAL"` no Cerbos) e transição como **ação nomeada**.

O conceito existe em outras camadas, o que mostra que é útil:
- em **bibliotecas de aplicação**;
- em **workflow/ERP** (nível 3, fora do nosso escopo).

Ressalva importante: a capacidade é **expressável** nos motores atuais (status como atributo + ação = transição). O diferencial está:
- na **definição única** de estados e transições, validada e compartilhada com a aplicação via SDK;
- na **checagem de legalidade da transição** (a transição T existe a partir do estado S?);
- em políticas e aprovação M de N **por transição**;
- na auditoria de transições e na UI administrativa.

Não está em algo impossível de outra forma. Por isso o diferencial pode ser copiado relativamente rápido.

### Evidência

| Onde | O que existe | Fonte |
|---|---|---|
| Cerbos | Estado como atributo em condição; o PlanResources gera filtro `status == "PENDING_APPROVAL"` para a ação `approve`. É stateless: "relies entirely on the application to provide… resource attributes" | docs.cerbos.dev/cerbos/latest/api; sapl.io/guides/comparison (comparativo de concorrente) |
| OpenFGA | Estado só como condição CEL; nada de transições | `openfga.md` |
| Axiomatics, PlainID, Keycloak AS | Estado como atributo; nenhum conceito de transição | arquivos deste diretório |
| Oso (Polar) | Recursos, relações e fatos; nenhum exemplo ou conceito de estado/transição na referência | osohq.com/docs (resource blocks, modeling) |
| Spring Statemachine | **Transições protegidas** declarativamente (`.secured(...)`, atributos `TRANSITION_SOURCE_`/`TRANSITION_TARGET_`, expressões, votação ANY/ALL/MAJORITY) | docs.spring.io/spring-statemachine (sm-security) |
| django-fsm | `@transition(..., permission=...)` e `has_transition_perm(obj.method, user)` | github.com/viewflow/django-fsm |
| Rails (`transition_buttons`) | Mostra só eventos legais **e** autorizados; checa de novo no servidor | rubygems.org/gems/transition_buttons |
| Oracle Siebel (state model) | Transição com autorização própria; por padrão, quem pode atualizar pode mudar o estado | docs.oracle.com |
| ADR público (CivitasConnect) | OPA para autorizar e máquina de estados no backend; "cada transição = um endpoint, uma permissão" | docs.core.civitasconnect.digital (ADR 032) |
| AuthZEN AARP | Usa máquina de estados só para a própria tarefa de aprovação (pendente → terminal) | openid.github.io/authzen |
| Issue pública (agent-plugins-skills #634) | Pede aprovação humana vinculada a **uma transição exata** (Auth0 CIBA + RAR), porque "existing controls only prove a human approval record exists" | github.com/richfrem/agent-plugins-skills/issues/634 |

Busca dirigida ("authorization product declarative state machine transitions") não retornou produto de AuthZ com esse recurso; só padrões e bibliotecas.

### Recomendação
- Manter o nível 2 (seção 4.29) como diferencial de camada posterior.
- A biblioteca de máquinas de estados (nível 1) deve ser o mesmo modelo usado pelo AuthZ, para garantir a definição única.
- Mapear no contrato AuthZEN como ação = transição, com o estado em `resource.properties` e o modelo de transições do lado do PDP. Assim continua compatível com qualquer PEP AuthZEN.
- A **combinação (a)+(b)**, ou seja, "transição T no estado S exige M de N com assinatura", não foi encontrada em nenhum lugar. É o argumento mais forte.

## Fora da lista pedida: o que foi visto
| Player | Relevância para (a)/(b) | Fonte |
|---|---|---|
| Cerbos | (a) não; (b) só atributo | docs.cerbos.dev |
| Oso | (a) não encontrado; (b) não | osohq.com/docs |
| OpenFGA | (a) não; (b) não | `openfga.md` |
| Permit.io | (a) aprovação de operação única, sem quórum documentado; (b) não | docs.permit.io |
| Indykite (KBAC) | Grafo de conhecimento + AuthZEN + MCP; aprovação e estados não encontrados | kuppingercole.com, skills.sh/indykite (fontes secundárias) |
| Styra | Fundadores do OPA e engenheiros foram para a Apple (08/2025); destino do DAS não esclarecido; nada sobre (a)/(b) | openpolicyagent.org blog; cloudnativenow.com |
| SGNL | Autorização contínua/JIT; aquisição pela CrowdStrike anunciada em 01/2026 (~US$ 740 mi); aprovação M de N não verificada | ir.crowdstrike.com; csoonline.com |
| ConductorOne, Opal, Lumos | (a) estágios sequenciais, sem M de N verdadeiro; (b) não | docs oficiais |
| Turnkey, Fireblocks | (a) sim (quórum); Turnkey com passkey sobre o corpo exato | docs oficiais |
| ScrambleID Lockstep | (a) mesmo conceito, **não lançado** | scrambleid.com |
| AuthZEN AARP | Envelope padronizado; quórum e assinaturas ficam fora de escopo | openid.github.io/authzen |

## Não verificado
- Se a Teleport vincula o desafio WebAuthn ao conteúdo da revisão. Documentação e RFD não dizem; o código não foi lido.
- Detalhes de aprovação do Permit.io (Operation Approval com vários aprovadores).
- Approval flows de Indykite, SGNL e Styra DAS.
- Data de lançamento do ScrambleID Lockstep.

## Fontes
- https://goteleport.com/docs/identity-governance/access-requests/access-request-configuration.md — 2026-11-03 — oficial
- https://goteleport.com/docs/zero-trust-access/authentication/mfa-for-admin-actions.md — 2026-11-03 — oficial
- https://raw.githubusercontent.com/gravitational/teleport/master/rfd/0131-adminitrative-actions-mfa.md — 2026-11-03 — RFD oficial
- https://developer.hashicorp.com/vault/docs/enterprise/control-groups — 2026-11-03 — oficial
- https://cloud.google.com/iam/docs/pam-create-entitlements — 2026-11-03 — oficial
- https://docs.aws.amazon.com/mpa/latest/userguide/mpa-concepts.html — 2026-11-03 — oficial
- https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-approval-workflow — 2026-11-03 — oficial
- https://docs.azure.cn/en-us/entra/id-governance/entitlement-management-access-package-approval-policy — 2026-11-03 — oficial
- https://kb.wisc.edu/security/115012 — 2026-11-03 — KB de cliente CyberArk
- https://docs.gitlab.com/ci/environments/deployment_approvals/ — 2026-11-03 — oficial
- https://developers.fireblocks.com/reference/transaction-authorization-objects.md — 2026-11-03 — oficial
- https://docs.turnkey.com/developer-reference/api-overview/stamps.md — 2026-11-03 — oficial
- https://docs.turnkey.com/concepts/policies/overview.md — 2026-11-03 — oficial
- https://docs.turnkey.com/api-reference/activities/approve-activity.md — 2026-11-03 — oficial
- https://turnkey.com/blog/introducing-verifiable-policy-decisions — 2026-11-05 — blog do fornecedor
- https://www.scrambleid.com/learn/scrambleid-lockstep-dual-control — 2026-11-05 — fornecedor (produto em desenvolvimento)
- https://knowledge.workspace.google.com/admin/security/multi-party-approval-for-sensitive-actions — 2026-11-03 — oficial
- https://learn.microsoft.com/en-us/intune/fundamentals/role-based-access-control/multi-admin-approval — 2026-11-03 — oficial
- https://docs.azure.cn/en-us/backup/multi-user-authorization-concept — 2026-11-03 — oficial
- https://docs.permit.io/overview/access-requests-and-approvals — 2026-11-03 — oficial
- https://www.c1.ai/docs/product/admin/policies.md — 2026-11-03 — oficial (ConductorOne)
- https://docs.opal.dev/docs/configure-reviewers — 2026-11-03 — oficial
- https://support.lumos.com/articles/7910043985-lumos-appstore-quick-start-guide — 2026-11-03 — oficial
- https://help.okta.com/en-us/Content/topics/identity-governance/access-requests/rcar-conditions.htm — 2026-11-03 — oficial
- https://openid.net/wg/authzen/specifications/ — 2026-11-03 — OpenID Foundation
- https://openid.github.io/authzen/authzen-access-request-approval-profile-1_0.html — 2026-11-03 — rascunho OpenID (Draft 1)
- https://github.com/w3c/webauthn/issues/2022 — 2026-11-03 — W3C (transaction confirmation)
- https://w3c.github.io/secure-payment-confirmation — 2026-11-03 — W3C
- https://docs.cerbos.dev/cerbos/latest/api/ — 2026-11-03 — oficial
- https://sapl.io/guides/comparison/ — 2026-11-03 — comparativo publicado por concorrente (viés)
- https://www.osohq.com/docs/reference/resource-blocks — 2026-11-03 — oficial
- https://docs.spring.io/spring-statemachine/docs/1.2.13.RELEASE/reference/html/sm-security.html — 2026-11-03 — oficial (versão antiga)
- https://github.com/viewflow/django-fsm — 2026-11-03 — repositório
- https://rubygems.org/gems/transition_buttons — 2026-11-03 — pacote
- https://docs.oracle.com/cd/E05554_01/books/BizProcessFINS/BizProcessUseStateMachine2.html — 2026-11-03 — oficial Oracle
- https://docs.core.civitasconnect.digital/docs_v2/next/Architecture/Architecture_Decisions/adrs_to_be_reviewed/adr032/ — 2026-11-03 — ADR público
- https://github.com/richfrem/agent-plugins-skills/issues/634 — 2026-11-03 — issue pública
- https://www.openpolicyagent.org/blog/rss.xml — 2026-11-03 — blog OPA (Styra/Apple)
- https://ir.crowdstrike.com/news-releases/news-release-details/crowdstrike-acquire-sgnl-transform-identity-security-ai-era — 2026-11-03 — press release
- https://ab-www.kuppingercole.com/research/ev81230/indykite-knowledge-based-access-control-kbac — 2026-11-03 — analista
