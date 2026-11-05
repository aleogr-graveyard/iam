# Aprovações e acesso privilegiado (referência para M de N)

Teleport (access requests) · HashiCorp Vault (control groups) · GCP Privileged Access Manager · AWS (TEAM e Multi-party approval) · Microsoft Entra PIM · outros com M de N (Google Workspace, Intune, Azure Backup MUA, CyberArk, GitLab, Fireblocks, Turnkey, IGAs)

| Campo | Valor |
|---|---|
| Categoria | PAM / JIT / aprovações de acesso e de operações sensíveis |
| Site / repositório | Teleport: https://goteleport.com/docs/identity-governance/access-requests/ · Vault: https://developer.hashicorp.com/vault/docs/enterprise/control-groups · GCP: https://cloud.google.com/iam/docs/pam-create-entitlements · AWS MPA: https://docs.aws.amazon.com/mpa/latest/userguide/ · AWS TEAM: https://github.com/aws-samples/iam-identity-center-team · Entra PIM: https://learn.microsoft.com/entra/id-governance/privileged-identity-management/ |
| Licença e modelo de negócio | Teleport: regras de aprovação são **Enterprise** (doc oficial). Vault: control groups exigem **Vault Enterprise ou HCP Vault Dedicated**. GCP PAM: 2 níveis e número de aprovações exigem **Security Command Center Premium/Enterprise** (preview). AWS MPA: recurso do AWS Organizations (exige IAM Identity Center). AWS TEAM: amostra open source (aws-samples), "AWS Content", não produto suportado. Entra PIM: licença P2/ID Governance |
| Stack | Teleport e Vault: Go. Nuvens: SaaS |
| Maturidade | Todos em produção; recursos de M de N mais recentes: AWS MPA (2025; baselining em 03/2026), GCP PAM 2 níveis (preview) |
| Acesso | 2026-11-03 |

## Resumo
- **M de N existe no mercado de PAM**, com níveis variados: Teleport (thresholds com filtros por papel/atributo do revisor), Vault (fatores com N aprovações por grupo, AND entre fatores), GCP PAM (até 5 aprovações por nível, 2 níveis), AWS MPA (M de N obrigatório, ≥2 de ≥3), CyberArk (número de confirmadores), GitLab (aprovações de deploy por regra).
- **Entra PIM é 1 de N**: "resolvido pelo primeiro aprovador"; multietapas não suportado.
- **Assinatura WebAuthn vinculada à ação exata**, por aprovador, com quórum: encontrada só em infraestrutura de carteiras cripto (**Turnkey**: desafio WebAuthn = SHA-256 do corpo da requisição; consenso `approvers…count() >= 2`). Teleport exige WebAuthn por requisição para aprovar, mas não há evidência de vínculo ao conteúdo.
- O rascunho **AuthZEN Access Request and Approval Profile** (Draft 1, 29/10/2026) padroniza o envelope (pedido, tarefa, reavaliação) e exige checagem de elegibilidade/SoD, mas deixa **fora de escopo** linguagens de política de aprovação e motores de workflow.

## Funcionalidades por domínio
Foco em 5.2 domínios 5 e 10; os demais domínios não são o objeto destes produtos.

| Domínio (5.2) | Teleport | Vault CG | GCP PAM | AWS MPA/TEAM | Entra PIM | Notas |
|---|---|---|---|---|---|---|
| 1. Modelos | ◐ | ◐ | ◐ | ○ | ◐ | RBAC de infraestrutura (papéis, políticas de caminho, papéis IAM) |
| 2. Decisões | ○ | ○ | ○ | ○ | ○ | Não são PDPs de aplicação |
| 3. Políticas | ◐ | ◐ | ◐ | ○ | ◐ | Teleport: linguagem de predicados nos filtros; Vault: HCL/Sentinel |
| 4. Dados | ○ | ○ | ○ | ○ | ○ | — |
| 5. Acesso dinâmico | ● | ● | ● | ◐ | ● | JIT + aprovação em todos; M de N: Teleport, Vault, GCP, AWS MPA; Entra 1 de N |
| 6. Estados e transições | ○ | ○ | ○ | ○ | ○ | Só o ciclo de vida do próprio pedido (pendente → aprovado/negado/expirado) |
| 7. Padrões | ○ | ○ | ○ | ○ | ○ | AuthZEN AARP: nenhum declara suporte (rascunho de 10/2026) |
| 8. Máquinas e agentes | ◐ | ◐ | ◐ | ○ | ◐ | GCP: service accounts/agent identities como aprovadores (preview); Teleport: revisões automáticas por regra |
| 9. Desempenho e resiliência | ? | ? | ? | ? | ? | Não aplicável/não publicado |
| 10. Auditoria | ● | ● | ● | ● | ● | Trilhas de pedido, revisão e uso em todos |
| 11. Experiência e administração | ● | ◐ | ● | ◐ | ● | UIs, Slack/e-mail (Teleport plugins), portal de aprovação (AWS) |

## Comparativo de aprovação (o que importa para o M de N do projeto)

| Produto | Quórum | Composição por papéis | Autoaprovação | Uma aprovação por pessoa | Janela/expiração | Forte autenticação por aprovação | Vínculo à ação exata |
|---|---|---|---|---|---|---|---|
| Teleport access requests | `approve: N`, `deny: N` por threshold; vários thresholds (OR) | Sim: `filter` com `reviewer.roles`, `reviewer.traits`, `request.reason` | Proibida | ? | Pedido com TTL | **WebAuthn por requisição** quando "MFA for admin actions" está ativo (aprovar pedido é ação administrativa) | Pedido = conjunto de papéis/recursos; MFA não documentado como ligado ao conteúdo |
| Vault control groups | `approvals = N` por fator; vários fatores (AND) | Sim: fator por grupo de identidade | ? | Pessoa em 2 grupos satisfaz os 2 fatores com 1 autorização | `ttl` do control group | Não encontrado | **Sim, por estado no servidor**: requisição original fica embrulhada (wrapping token) e só é executada após as autorizações |
| GCP PAM | `approvalsNeeded` até 5 por nível, até 2 níveis sequenciais | Por lista de aprovadores por nível | Proibida | "Pode aprovar só uma vez" mesmo em 2 níveis | 24 h para decidir | Não encontrado | Grant = papéis + escopo + duração |
| AWS Multi-party approval | M de N; mínimo 2 de 3 | Não (uma equipe) | ? | ? | Sessão de 24 h; expirado = rejeição | Login IAM Identity Center | **Operação protegida** específica (ex.: `CreateRestoreAccessVault`); executada automaticamente com permissões do solicitante |
| AWS TEAM | Aprovação por grupo de aprovadores (quórum >1 não encontrado) | Por elegibilidade | ? | ? | Duração do acesso | Login IAM Identity Center | Conta + permission set + duração |
| Entra PIM | **1 de N** ("primeiro aprovador resolve") | Não | — | — | 24 h fixas | MFA na ativação (do solicitante) | Papel + escopo |
| Entra entitlement mgmt | até 3 estágios, **1 por estágio** | Por estágio | ? | ? | Por política | — | Pacote de acesso |
| Google Workspace MPA | 2 pessoas (outro super admin aprova) | Privilégio de revisor | Proibida | — | 72 h (imprensa) | — | Alteração de configuração específica |
| Intune Multi Admin Approval | 2 pessoas | Grupo aprovador por política | Proibida, inclusive Global Admin | — | ? | — | Alteração específica (apps, scripts, wipe) |
| Azure Backup MUA | 2 papéis (dono do Resource Guard concede via PIM) | Separação backup admin × security admin | — | — | Via PIM | — | Operações críticas protegidas |
| CyberArk dual control | Número de confirmadores na Master Policy | ? | ? | ? | ? | ? | Acesso a conta privilegiada |
| GitLab deployment approvals | `required_approvals` por regra, várias regras | Sim, por grupo/papel | Bloqueada por padrão | Sim, uma por deploy | ? | — | Job de deploy específico |
| Fireblocks | Threshold por grupo, AND/OR entre grupos | Sim | Opcional (iniciador conta) | ? | ? | Dispositivo móvel do aprovador | Transação |
| **Turnkey** | `consensus` em expressão (`approvers.filter(...).count() >= 2`) | Sim (tags de usuário) | ? | ? | Janela de validade da requisição assinada | **WebAuthn/passkey** (ou API key) por requisição | **Sim, criptográfico**: desafio WebAuthn = SHA-256 do corpo; aprovação por `fingerprint` da atividade; verificado em enclave |
| Opal / ConductorOne / Lumos / Okta OIG | Estágios sequenciais (Opal: All/Any por estágio; C1: grupo = 1 membro basta) | Por estágio | ? | ? | Escalonamento | — | Pedido de acesso (IGA) |

## Diferenciais / exclusivos
- **Teleport:** quórum com filtros expressivos (ex.: "3 aprovações de quem não é do time dev, ou 1 de admin"), regras de negação simétricas, revisões automáticas que respeitam thresholds.
- **Vault:** a requisição pendente **é** a ação; aprovação não pode ser reaproveitada para outra operação.
- **AWS MPA:** equipes "autoprotegidas" (mudar a equipe exige aprovação da equipe), baselining para testar a equipe, aprovador inativo após 2 sessões sem resposta.
- **Intune MAA:** a própria criação de política de aprovação exige segundo administrador.
- **Turnkey:** único caso encontrado de quórum + passkey + assinatura sobre o conteúdo exato + avaliação de política em enclave.

## O que a comunidade elogia (com evidência)
- Avaliações sistemáticas não coletadas (limite de buscas da sessão). Sinal indireto: grandes provedores lançaram M de N recentemente (AWS MPA 2025, Workspace 2024/2025, GCP 2 níveis em preview), o que indica demanda.
- Blog japonês (Classmethod) documenta um comportamento contraintuitivo do AWS MPA: aprovação vale mesmo com maioria contra, se o limiar de "sim" for atingido — mostra a importância de regra de negação explícita (Teleport tem `deny: N`).

## Do que a comunidade se queixa (com evidência)
- Entra PIM: sem multietapas, 24 h não configuráveis, primeiro aprovador decide (doc oficial — limitação, não queixa coletada).
- GCP PAM: M de N e 2 níveis atrás de licença SCC Premium/Enterprise e em preview; a referência do Config Connector ainda diz "Only 1 step is supported" (inconsistência de documentação).
- Teleport: regras de aprovação só no Enterprise; MFA para ações administrativas quebra automação com `tctl auth sign` (doc oficial).
- Vault: control groups só no Enterprise; ferramentas OSS não têm equivalente.
- AWS TEAM: é amostra, "você é responsável por testar, proteger e otimizar" (README).

## Histórico de segurança
- Não houve tempo/buscas para levantar CVEs específicos de fluxos de aprovação destes produtos (limite de buscas). **Lições de design extraídas da documentação:**
  - Teleport RFD 0131 rejeitou certificados "MFA-verificados" de 1 minuto porque **uma MFA autorizaria várias ações** (ataque de replay de MFA); escolheu MFA por requisição.
  - Vault: aprovar o **pedido armazenado** (não um token genérico) impede usar a aprovação para outra operação.
  - AWS MPA: rejeição implícita por expiração e por não resposta; mudanças na equipe protegidas pela própria equipe.
  - AuthZEN AARP: "aprovação não é bearer grant"; o PDP reavalia política atual e verifica vínculo a sujeito/recurso/ação/contexto/expiração.

## Arquitetura e decisões notáveis
- Dois padrões de execução: (1) **conceder acesso temporário** e o usuário age depois (Teleport, GCP PAM, Entra PIM, TEAM); (2) **segurar a operação e executá-la ao atingir o quórum** (Vault, AWS MPA, Workspace, Intune). O padrão 2 é o que corresponde a "aprovação vinculada à ação exata".
- AARP adota o padrão 1 com reavaliação: depois da aprovação, o PEP refaz a avaliação AuthZEN com `context.approval`.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Nuvens: aprovações restritas ao próprio provedor e a operações predefinidas (AWS MPA: lista fixa de operações protegidas).
- Teleport/Vault: self-hosted possível; extensibilidade por plugins (Slack, PagerDuty, Jira) e API.
- Nenhum oferece aprovação M de N como **serviço genérico para ações de aplicações de terceiros** (exceto Turnkey, restrito a carteiras, e Permit.io "Operation Approval", sem quórum documentado).

## (Somente AuthZ) Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- Não aplicável na maior parte (não são PDPs). AuthZEN: nenhum declara suporte ao AARP; o AARP é o ponto de encaixe natural do nosso fluxo de aprovação.

## Verificação de diferenciais candidatos do projeto
- **Aprovação M de N de primeira classe?** **Sim, em PAM** (Teleport, Vault Enterprise, GCP PAM, AWS MPA, CyberArk, GitLab, Fireblocks, Turnkey). **Não** em Entra PIM (1 de N). Portanto M de N **sozinho não é diferencial**; ver `differentials-check.md` para a combinação.
- **Autorização sensível ao estado / transições?** **Não** em nenhum. Só o ciclo de vida do pedido de aprovação é uma máquina de estados.

## Implicações para o projeto
1. **Adotar:** thresholds com regra de aprovação **e** regra de negação (`approve: M`, `deny: K`), com filtros por papel/atributo do aprovador (modelo Teleport) — cobre "1 da segurança e 1 do financeiro".
2. **Adotar:** padrão "segurar a operação" (Vault/AWS MPA) para ações críticas: o que é aprovado é a representação canônica da ação pendente, executada só no quórum.
3. **Adotar:** uma aprovação por pessoa mesmo em vários níveis (GCP), proibição de autoaprovação, expiração = rejeição (AWS), e proteção da própria política de aprovação por aprovação (AWS MPA, Intune).
4. **Adotar:** MFA/WebAuthn **por ação**, nunca sessão "elevada" reaproveitável (lição do RFD 0131 da Teleport).
5. **Oportunidade:** nenhum produto de identidade assina a ação exata por aprovador com WebAuthn; Turnkey prova que é viável (desafio = hash da requisição).
6. **Adotar:** implementar o envelope do AuthZEN AARP (pedido após negação requisitável, tarefa, reavaliação) e colocar o M de N **dentro** do Access Request Service, que o AARP deixa fora de escopo.

## Fontes
- https://goteleport.com/docs/identity-governance/access-requests/access-request-configuration.md — 2026-11-03 — oficial
- https://goteleport.com/docs/zero-trust-access/authentication/mfa-for-admin-actions.md — 2026-11-03 — oficial
- https://raw.githubusercontent.com/gravitational/teleport/master/rfd/0131-adminitrative-actions-mfa.md — 2026-11-03 — RFD oficial
- https://goteleport.com/docs/ver/19.x/identity-governance/access-requests/oss-role-requests/ — 2026-11-03 — oficial (limites da edição Community)
- https://developer.hashicorp.com/vault/docs/enterprise/control-groups — 2026-11-03 — oficial
- https://cloud.google.com/iam/docs/pam-create-entitlements — 2026-11-03 — oficial
- https://cloud.google.com/iam/docs/pam-approve-deny-grants — 2026-11-03 — oficial
- https://cloud.google.com/config-connector/docs/reference/resource-docs/privilegedaccessmanager/privilegedaccessmanagerentitlement — 2026-11-03 — oficial
- https://docs.aws.amazon.com/mpa/latest/userguide/mpa-concepts.html — 2026-11-03 — oficial
- https://docs.aws.amazon.com/mpa/latest/userguide/what-is.html — 2026-11-03 — oficial
- https://docs.aws.amazon.com/mpa/latest/userguide/create-team.html — 2026-11-03 — oficial
- https://aws.amazon.com/about-aws/whats-new/2026/03/multi-party-approval-team-baselining — 2026-11-03 — oficial
- https://dev.classmethod.jp/articles/mpa-min-number-of-people/ — 2026-11-03 — blog de parceiro AWS
- https://raw.githubusercontent.com/aws-samples/iam-identity-center-team/main/README.md — 2026-11-03 — oficial (amostra)
- https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-approval-workflow — 2026-11-03 — oficial
- https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/groups-approval-workflow — 2026-11-03 — oficial
- https://docs.azure.cn/en-us/entra/id-governance/entitlement-management-access-package-approval-policy — 2026-11-03 — oficial
- https://knowledge.workspace.google.com/admin/security/multi-party-approval-for-sensitive-actions — 2026-11-03 — oficial
- https://workspaceupdates.googleblog.com/2025/06/more-granular-controls-for-multi-party-approvals-for-sensitive-admin-actions.html — 2026-11-03 — oficial
- https://helpnetsecurity.com/?p=297392 — 2026-11-03 — imprensa (72 h)
- https://learn.microsoft.com/en-us/intune/fundamentals/role-based-access-control/multi-admin-approval — 2026-11-03 — oficial
- https://docs.azure.cn/en-us/backup/multi-user-authorization-concept — 2026-11-03 — oficial
- https://kb.wisc.edu/security/115012 — 2026-11-03 — KB de universidade cliente (CyberArk)
- https://docs.gitlab.com/ci/environments/deployment_approvals/ — 2026-11-03 — oficial
- https://developers.fireblocks.com/reference/transaction-authorization-objects.md — 2026-11-03 — oficial
- https://docs.turnkey.com/developer-reference/api-overview/stamps.md — 2026-11-03 — oficial
- https://docs.turnkey.com/concepts/policies/overview.md — 2026-11-03 — oficial
- https://docs.turnkey.com/api-reference/activities/approve-activity.md — 2026-11-03 — oficial
- https://docs.opal.dev/docs/configure-reviewers — 2026-11-03 — oficial
- https://www.c1.ai/docs/product/admin/policies.md — 2026-11-03 — oficial (ConductorOne)
- https://support.lumos.com/articles/7910043985-lumos-appstore-quick-start-guide — 2026-11-03 — oficial
- https://help.okta.com/en-us/Content/topics/identity-governance/access-requests/rcar-conditions.htm — 2026-11-03 — oficial
- https://docs.permit.io/overview/access-requests-and-approvals — 2026-11-03 — oficial
- https://openid.github.io/authzen/authzen-access-request-approval-profile-1_0.html — 2026-11-03 — rascunho OpenID (Draft 1, 29/10/2026)
