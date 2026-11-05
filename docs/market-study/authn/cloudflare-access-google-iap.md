# Cloudflare Access e Google Identity-Aware Proxy (IAP) — comparativo

| Campo | Cloudflare Access | Google IAP |
|---|---|---|
| Categoria | Proxy de identidade gerenciado (ZTNA) na borda da rede Cloudflare; parte do Cloudflare One (SASE) | Camada de autorização gerenciada no Google Cloud (Load Balancing, Cloud Run, App Engine, GKE, VMs via TCP) |
| Site | https://developers.cloudflare.com/cloudflare-one/access-controls/ | https://cloud.google.com/iap/docs/concepts-overview |
| Licença e modelo de negócio | Proprietário (SaaS). Zero Trust grátis até 50 usuários; pay-as-you-go US$ 7/usuário/mês; contrato Enterprise. Conector `cloudflared` é open source | Proprietário. IAP **sem custo** para recursos no Google Cloud (paga-se load balancer/compute). Pagos via Chrome Enterprise Premium: proxy de recursos fora do GCP, customização, atributos de dispositivo e níveis de acesso customizados. Identidades externas cobradas pelo Identity Platform |
| Stack | Rede anycast global da Cloudflare; Workers/Workers KV em partes do plano de controle (ver incidente de jun/2025); `cloudflared` (Go) para túneis | Infraestrutura de serving do Google (GFE/Load Balancing); IAM do Google Cloud; Access Context Manager |
| Maturidade | Lançado jan/2018 ("Like BeyondCorp, but you don't have to be at Google", [HN 16167964](https://news.ycombinator.com/item?id=16167964), 116 pts) | Lançado mar/2017, versão pública do BeyondCorp ([HN 13833855](https://news.ycombinator.com/item?id=13833855), 196 pts) |
| Acesso | 2026-11-03 | 2026-11-03 |

> Limitações: produtos fechados, sem repositório nem issues públicas; evidência de comunidade limitada a HN e pesquisas de segurança. Orçamento de buscas web esgotado no fim da pesquisa.

## Resumo
- Ambos implementam o modelo BeyondCorp: aplicação sem VPN, cada requisição autenticada no proxy, identidade repassada ao upstream em **JWT assinado** (`Cf-Access-Jwt-Assertion` / `x-goog-iap-jwt-assertion`).
- **Cloudflare Access** é agnóstico de nuvem e de IdP (vários IdPs simultâneos, OTP por e-mail), com MFA independente do IdP, tokens de serviço, mTLS, SCIM, aprovação temporária, OAuth gerenciado para agentes e portais MCP.
- **Google IAP** é amarrado ao Google Cloud e ao IAM; brilha pela integração (sem custo, sem infra) e por identidades externas via Identity Platform; contexto avançado é pago.
- Nenhum dos dois é IdP completo para usuários finais; ambos dependem de "a origem só aceitar tráfego do proxy" — origem exposta = bypass.

## Funcionalidades por domínio (5.1)

| # | Domínio | Cloudflare Access | Google IAP |
|---|---|---|---|
| 1 | Autenticação primária | ◐ IdPs múltiplos simultâneos (sociais, corporativos, OIDC/SAML genéricos), IdP "Cloudflare account" padrão, **OTP por e-mail** sem configuração. Sem senha própria nem LDAP direto | ◐ Contas Google; Workforce Identity Federation (IdP externo); Identity Platform (e-mail/senha, social, SAML/OIDC) para clientes |
| 2 | MFA e step-up | ◐ **Independent MFA** por aplicação/política: TOTP, chaves WebAuthn, biometria de plataforma; PIV/FIDO2 YubiKey para SSH. Sem SMS/push | ◐ Herdado da conta Google / Identity Platform (SMS, TOTP). Step-up por política: não encontrado |
| 3 | Ciclo de vida da conta | ◐ Provisionamento SCIM (de IdP para Cloudflare), gestão de assentos; sem cadastro de usuário final | ○ Inexistente no IAP (fica no Cloud Identity/Identity Platform) |
| 4 | Sessões e tokens | ◐ Dois JWTs (token global no team domain + token por aplicação); duração por app/política; revogação por app ou por usuário (efetiva em até ~1 min) | ◐ Cookie de sessão após login; gestão de sessões de identidades externas documentada; revogação granular: não encontrado |
| 5 | Protocolos | ◐ RP OIDC/SAML (AuthN requests SAML assinadas); **IdP para SaaS** (OIDC/SAML, catálogo de ~35 apps); **Managed OAuth** (Access como servidor OAuth 2.0, exige RFC 8707) | ◐ RP via Google/Identity Platform; JWT assinado ao upstream; integração SAML (propagação de atributos SAML) |
| 6 | Máquinas e agentes | ● Service tokens (Client ID/Secret, rotação, expiração, desativação de inativos), mTLS, `cloudflared access`, guia para agentes de código, OAuth gerenciado e portais MCP | ◐ Autenticação programática com tokens OIDC de contas de serviço; Agent Gateway (Gemini Enterprise Agent Platform) aplica IAM via IAP |
| 7 | Detecção e defesa | ◐ Postura de dispositivo (cliente Cloudflare One, MDMs), risk score de usuário, integração com risco do Entra ID; WAF/bot da plataforma | ◐ Context-aware access (IP, geografia; dispositivo e níveis customizados pagos) |
| 8 | Privacidade e conformidade | ◐ Logs de acesso, gestão de PII em logs, Logpush; localização de dados: não verificado | ◐ Cloud Audit Logging; certificações do GCP (Identity Platform: ISO 27001, SOC, BAA) |
| 9 | Experiência e administração | ◐ Dashboard, App Launcher, páginas de bloqueio, Terraform/API | ◐ Console GCP, IAM; customização da tela é paga (CEP) |
| 10 | Operação e migração | ◐ SaaS; túneis `cloudflared` com réplicas; depende da disponibilidade da Cloudflare | ◐ SaaS; habilitar por backend/serviço; depende do GCP |
| 11 | Identidade descentralizada | ○ | ○ |

## Diferenciais / exclusivos
**Cloudflare Access**
- **Temporary authentication** com justificativa: usuário pede acesso, aprovadores (lista de e-mails, autenticados pelo Access) aprovam por até 24 h; logs JIT via API com status PENDING/APPROVED/DENIED/CANCELED e "approval count".
- **Independent MFA**: exigir chave física numa aplicação sensível sem mudar o IdP.
- **Managed OAuth**: transforma qualquer app protegida em recurso OAuth para CLIs e agentes, com as mesmas políticas do login no navegador.
- **Portais MCP** e autenticação de agentes de código (`cloudflared access curl`, service tokens).
- **Access for SaaS**: o proxy vira IdP SAML/OIDC para SaaS de terceiros, aplicando as mesmas políticas.

**Google IAP**
- **Grátis** para recursos no GCP e integrado ao IAM (papel "IAP-secured Web App User"), com IAM Conditions.
- **TCP forwarding** para SSH/RDP sem IP público.
- **Identidades externas** via Identity Platform: o mesmo proxy serve funcionários e clientes.
- No ecossistema GCP, o **Privileged Access Manager** (PAM, produto separado) oferece elevação JIT com aprovação **multi-party/multi-level** (Preview).

## O que a comunidade elogia (com evidência)
- Ambos popularizaram o fim da VPN para apps internas; comentários de 2017/2018 elogiam o modelo BeyondCorp e pedem equivalentes em outras nuvens ([HN 13833855](https://news.ycombinator.com/item?id=13833855), [HN 16167964](https://news.ycombinator.com/item?id=16167964)).
- IAP: "muito mais barato" que alternativas como Azure AD App Proxy (mesmo thread 13833855).
- Cloudflare: plano grátis até 50 usuários torna o ZTNA acessível para pequenas equipes (preço oficial via busca; agregadores de preço — fontes secundárias).

## Do que a comunidade se queixa (com evidência)
- **Confiança e concentração**: "granting Cloudflare access to all resources" e desconfiança pós-Cloudbleed ([HN 16167964](https://news.ycombinator.com/item?id=16167964), 2018).
- **Só HTTP** no início ("all systems are not HTTP"; mesmo thread) — mitigado depois com túneis TCP/SSH/RDP.
- **"Fachada de autenticação com logs"**: BeyondCorp real exigiria reputação de dispositivo, 2FA, geofencing (mesmo thread) — hoje parcialmente pago (postura, níveis de acesso).
- **Dependência do provedor**: Cloudflare Access caiu junto com Workers KV em 12/jun/2025 (~2h28; WARP não matriculava; o "emergency disconnect" também dependia do KV) ([postmortem](https://blog.cloudflare.com/cloudflare-service-outage-june-12-2025/)).
- **IAP**: recursos "de verdade" (dispositivo, níveis customizados, recursos fora do GCP) atrás do Chrome Enterprise Premium (página de preços); lock-in no GCP.

## Histórico de segurança
**Google IAP**
- **Bypass por compartilhamento de OAuth client (2021)**: atacante cria outra app IAP com o mesmo client ID e roteia, por query string `gcp-iap-mode=AUTHENTICATING`, para backend sem IAP, capturando o `redirect_token` da vítima e sequestrando a sessão na app alvo ([seblu.de](https://www.seblu.de/2021/12/iap-bypass.html), pesquisador; corrigido pelo Google).
- **Responsabilidade compartilhada documentada**: quem alcança a porta da VM/GKE contorna o IAP; com LB + Cloud Run, a URL automática do Cloud Run fica desprotegida (doc oficial).

**Cloudflare Access**
- Advisory específico de bypass do Access: **não encontrado** (buscas retornaram WAF/ACME e origin pull).
- **ACME path bypass do WAF** (relatado em out/2025 via HackerOne; corrigido; sem exploração conhecida) — classe "caminho especial isento de controles" ([securityaffairs](https://securityaffairs.com/187156/security/acme-flaw-in-cloudflare-allowed-attackers-to-reach-origin-servers.html) — imprensa).
- **Certificado compartilhado entre clientes no Authenticated Origin Pulls (2023)**: "vir da Cloudflare" não prova "vir do meu tenant" (Certitude; [techradar](https://www.techradar.com/pro/security/cloudflare-security-protections-can-be-bypassed-in-a-surprisingly-simple-way) — imprensa).

Lições para nós:
- **A origem deve validar o JWT assinado do proxy** (issuer, audience por aplicação, assinatura); restrição por rede não basta. Documentar e alertar quando o upstream não valida.
- **Audience/cliente por aplicação**: compartilhar OAuth client entre apps permitiu o bypass do IAP; vincular tokens de redirecionamento a uma app/rota específica.
- **Credenciais de confiança do proxy devem ser por tenant**, nunca compartilhadas entre clientes.
- **Caminhos especiais** (ACME, health check, callback) não podem pular controles.
- **Plano de controle não pode depender de um único armazenamento/terceiro**; o "kill switch" de emergência não pode depender do sistema que falhou (premissa 17).

## Arquitetura e decisões notáveis
- **Cloudflare**: token global (SSO entre apps, no team domain) + token por aplicação (no hostname da app); renovação silenciosa do token da app enquanto o global vale e a política continua satisfeita. Políticas Allow/Block/Bypass/Service Auth com ordem explícita. Aplicações self-hosted, SaaS, infraestrutura (SSH com certificados curtos) e redes privadas.
- **Google**: o serving (LB, Cloud Run, App Engine) consulta o servidor de autenticação do IAP; após login, checa o papel IAM no recurso; Context-Aware Access via Access Context Manager. IAP cria/gera um OAuth client por recurso (ou cliente gerenciado pelo Google).
- Ambos repassam identidade em JWT assinado com chaves públicas publicadas para verificação no upstream.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- **Multi-tenancy**: Cloudflare — uma organização Zero Trust por conta, várias IdPs; IAP — por projeto/recurso GCP; identidades externas multi-tenant via tenants do Identity Platform.
- **Extensibilidade**: Cloudflare — "external evaluation" e Workers (não verificado em detalhe), API, Terraform; IAP — IAM Conditions, Access Context Manager, API.
- **Configuração**: dashboard, API REST, Terraform (ambos).
- **Bancos**: não se aplica (SaaS).
- **Soberania**: dados de sessão/logs nos provedores; Cloudflare exige trafegar pela rede dela; IAP restrito ao GCP (exceto recursos externos pagos).

## Verificação de diferenciais candidatos do projeto
- Aprovação multipartes (M de N):
  - Cloudflare Access: **parcial**. Aprovação temporária com lista de aprovadores e campo "approval count" nos logs; a documentação não descreve quórum configurável (M de N) — interpretado como aprovação única.
  - Google IAP: **não** no IAP. O PAM do GCP (fora do IAP) tem aprovação multi-party/multi-level em **Preview** — sinal de que o mercado de nuvem começa a tratar quórum como recurso.
- Autorização sensível ao estado / transições declarativas: **não** em ambos (decidem por identidade, grupo, dispositivo, rede e tempo).

## Implicações para o projeto
1. **Adotar** o par "token global de SSO + token por aplicação com audience própria" e renovação silenciosa condicionada à política vigente.
2. **Adotar** MFA independente do IdP por aplicação/política (step-up na borda), inclusive com IdP externo federado.
3. **Adotar** "OAuth gerenciado" para clientes não-navegador e agentes (RFC 8707, metadados de recurso protegido), com as mesmas políticas do login web.
4. **Oportunidade**: aprovação JIT com quórum M de N nativa no AuthN/AuthZ — Cloudflare oferece aprovação simples; Google só no PAM em Preview.
5. **Evitar** dependência de um único provedor/armazenamento no caminho de login e no mecanismo de emergência; modo degradado documentado.
6. **Adotar** alertas de configuração para "origem acessível sem passar pelo proxy" e "upstream não valida JWT" (premissa 4).
7. **Oportunidade**: oferecer o mesmo modelo self-hosted e agnóstico de nuvem (premissas 9 e 19), o que nenhum dos dois entrega.

## Fontes
- https://developers.cloudflare.com/cloudflare-one/access-controls/policies/temporary-auth/index.md — 2026-11-03 — fornecedor
- https://developers.cloudflare.com/cloudflare-one/integrations/identity-providers/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/access-controls/access-settings/independent-mfa/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/access-controls/access-settings/session-management/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/access-controls/authenticate-agents/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/access-controls/applications/http-apps/managed-oauth/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/access-controls/service-credentials/service-tokens/index.md — 2026-11-03
- https://developers.cloudflare.com/cloudflare-one/llms.txt e .../access-controls/llms.txt — 2026-11-03
- https://www.cloudflare.com/plans/zero-trust-services/ e agregadores (costbench.com, spendhound.com) via busca — 2026-11-03 — preço: secundário
- https://blog.cloudflare.com/cloudflare-service-outage-june-12-2025/ (via resumo de busca) — 2026-11-03
- https://securityaffairs.com/187156/security/acme-flaw-in-cloudflare-allowed-attackers-to-reach-origin-servers.html — 2026-11-03 — imprensa
- https://www.techradar.com/pro/security/cloudflare-security-protections-can-be-bypassed-in-a-surprisingly-simple-way — 2026-11-03 — imprensa
- https://cloud.google.com/iap/docs/concepts-overview — 2026-11-03 — fornecedor
- https://cloud.google.com/iap/pricing — 2026-11-03
- https://cloud.google.com/iap/docs/agent-overview — 2026-11-03
- https://cloud.google.com/iam/docs/pam-overview — 2026-11-03
- https://www.seblu.de/2021/12/iap-bypass.html — 2026-11-03 — pesquisador independente
- HN via API Algolia: 16167964, 13833855, 42326458 — 2026-11-03
