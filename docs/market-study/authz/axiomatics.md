# Axiomatics (XACML, ALFA)

| Campo | Valor |
|---|---|
| Categoria | Plataforma corporativa de autorização ABAC/PBAC ("Authorization Management Platform"), referência histórica do XACML |
| Site / repositório | https://axiomatics.com · docs https://docs.axiomatics.com · linguagem https://alfa.guide · GitHub https://github.com/axiomatics (exemplos/ferramentas; motor fechado) |
| Licença e modelo de negócio | Proprietário, licença corporativa ("Appropriate license"), suporte pago. ALFA é linguagem publicada (guia público, playground, plugins VS Code/Eclipse); motor e ferramentas são comerciais. **Adquirida 100% pela Leonardo S.p.A. (Itália)**, anúncio de 09/07/2025, sujeita a aprovação FDI sueca (fechamento não confirmado nesta pesquisa) |
| Stack | Java (ADS em Spring Boot 4 desde 26.1; distribuído como JAR ou Helm/Kubernetes); XACML 3.0 como modelo interno; ALFA compila para XACML; JDBC para conectores; OpenTelemetry |
| Maturidade | Fundada em 2006 (Estocolmo). ADS versão 26.2.1 (versionamento anual desde 2026; ADS 2.x → 26.1). CTO (David Brossard) copreside o AuthZEN WG. Foco em defesa, governo, finanças, infraestrutura |
| Acesso | 2026-11-03 |

## Resumo
- Motor PDP XACML 3.0 (Access Decision Service, ADS) + control plane (Authorization Hub) + toolchain de política como código (Axiomatics Policy DevOps, APD, em Gradle/JUnit).
- **ALFA**: sintaxe enxuta que compila para XACML; resolve a principal queixa histórica do XACML (políticas em XML ilegíveis).
- **Reverse queries / CAQ**: pergunta "que condições levam a Permit?" — base para filtros de busca e mascaramento de UI (listagem).
- Desde ADS 26.2: **AuthZEN Authorization API 1.0** (avaliação, lote, descoberta), desligado por padrão.
- Não tem ReBAC nativo, solicitação/aprovação de acesso nem modelo de estados.

## Funcionalidades por domínio

| Domínio (5.2) | Suporte | Notas |
|---|---|---|
| 1. Modelos | ◐ | ABAC/PBAC completos; RBAC via atributos. ReBAC nativo: não (relações só via atributos buscados em conectores) |
| 2. Decisões | ● | Avaliação XACML (JSON/XML) com Multiple Decision Profile, AuthZEN single e batch, **reverse query** (CAQ, hoje dentro do ADS) para "quais recursos/condições". Explicabilidade: decisões Permit/Deny/NotApplicable/Indeterminate + status; "explain" detalhado não verificado |
| 3. Políticas | ● | ALFA; Authorization Hub (versionamento, ciclo de vida); APD com testes unitários, de integração e de sistema (JUnit 5) e pipelines CI/CD (Jenkins, Azure DevOps, OpenShift). Implantação gradual nativa: não encontrada |
| 4. Dados | ◐ | Attribute connectors (LDAP, SQL, HTTP/REST, parser) consultados na hora da decisão (PIP); sem armazém de relações; consistência = dado da fonte no momento |
| 5. Acesso dinâmico | ◐ | **Obrigações e advice** (XACML) devolvidas também no AuthZEN; regras de tempo/contexto. JIT, solicitações, aprovações, M de N, delegação administrativa: não encontrados |
| 6. Estados e transições | ○ | Estado só como atributo de recurso |
| 7. Padrões | ● | XACML 3.0, AuthZEN 1.0, rascunho IETF "AuthZEN profile for OAuth RAR" (Brossard/Axiomatics, expirado em jan/2025) |
| 8. Máquinas e agentes | ◐ | Sujeitos arbitrários; agenda de 2026 sobre AuthZEN para agentes/MCP (Identiverse 2026). Produto específico para MCP: não verificado |
| 9. Desempenho e resiliência | ◐ | ADS stateless, escala horizontal em K8s; pool JDBC configurável. Números de desempenho publicados: não encontrados |
| 10. Auditoria | ● | Logs de avaliação (request, decisão, obrigações); spans OTEL por chamada de conector; ADM com trilha para auditores |
| 11. Experiência e administração | ◐ | Authorization Hub (UI), plugins de IDE, playground ALFA; SDK/PEP prontos para várias linguagens: não verificado |

## Diferenciais / exclusivos
- **Reverse queries** (partial evaluation sobre XACML) para filtrar dados e montar UI — a capacidade de "listagem" do mundo ABAC.
- **Política como código madura**: projeto Gradle, testes JUnit em quatro níveis (unidade, integração com conectores, sistema, conector isolado), execução contra ADS real. Recomendação explícita de não depender de dados externos voláteis em testes automatizados.
- **Obrigações/advice** com semântica padronizada desde o XACML (o mesmo conceito que o AuthZEN agora padroniza no perfil de obrigações).
- Liderança em padrões: autoria do ALFA, coautoria do AuthZEN e do perfil RAR.

## O que a comunidade elogia (com evidência)
- Avaliações públicas não encontradas em quantidade (G2 tem página do produto; conteúdo não acessível nesta sessão).
- ALFA é recomendado por membros do TC XACML como "sintaxe muito mais amigável" (lista OASIS, 2016) — fonte antiga, viés de participante do comitê.

## Do que a comunidade se queixa (com evidência)
- Barreira histórica do XACML: "dificuldade de criar e modificar políticas" (lista XACML OASIS, 2016). ALFA mitiga, mas o modelo mental XACML (combining algorithms, targets, Indeterminate) continua complexo.
- Mudanças de dependência quebram upgrades: ADS 26.2.0 mudou salt de hash `{pbkdf2}` de 8 para 16 bytes por causa do Spring Security; hashes antigos falham (release notes oficiais).
- Bugs de logs de auditoria com JSON inválido (aspas em obrigações; atributos JSON aninhados) corrigidos em 26.x — mostra a fragilidade de logs montados por concatenação.

## Histórico de segurança
- NVD (keywordSearch "Axiomatics"): 0 CVEs. Página de divulgação de vulnerabilidades existe (axiomatics.com/vulnerability-disclosure-policy).
- No ecossistema XACML: CVE-2016-4312 (WSO2 Identity Server, XXE no fluxo XACML) — **lição:** formatos XML de política/requisição atraem XXE; preferir JSON e parsers sem entidades externas.
- **Lição dos bugs de auditoria:** serializar logs de decisão com encoder estruturado, nunca por concatenação.

## Arquitetura e decisões notáveis
- Três planos: controle (Authorization Hub), decisão (ADS), entrega (APD). "Authorization domain" é o artefato implantável (políticas + atributos + conectores), versionado e servido pelo ADM.
- ADS em modos `ads`, `caq` ou `ads_caq`; AuthZEN só nos modos com ADS.
- Mapeamento AuthZEN → atributos XACML padronizados (`subject.Id`, `resource.Type`, `action.name`…), com descoberta em `/.well-known/authzen-configuration`.
- Versionamento anual de releases (26.x) com compatibilidade retroativa declarada.

## Multi-tenancy, extensibilidade, configuração, bancos suportados, soberania de dados
- Multi-tenancy: por "authorization domains" separados e namespaces no dicionário de atributos; multi-tenancy de clientes finais não é conceito explícito.
- Extensibilidade: conectores de atributos e funções XACML customizadas (não verificado em detalhe).
- Bancos: conectores JDBC para fontes; armazenamento interno do Hub não publicado.
- Soberania: tudo pode rodar on-premises (JAR/K8s) — vantagem em defesa/governo; reforçada pela compra pela Leonardo.

## (Somente AuthZ) Modelos, linguagem, consistência, listagem, sincronização, explicabilidade, testes, avaliação local, desempenho, AuthZEN
- **Modelos:** ABAC/PBAC; RBAC por atributo.
- **Linguagem:** ALFA → XACML 3.0.
- **Consistência:** leitura de atributos na hora (sem cache de relações próprio), portanto consistente com a fonte, ao custo de latência.
- **Listagem:** reverse queries (CAQ).
- **Sincronização:** não sincroniza; consulta PIPs.
- **Explicabilidade:** status XACML, obrigações/advice; trilha de avaliação detalhada não verificada.
- **Testes:** APD/JUnit, quatro níveis.
- **Avaliação local:** ADS como serviço local/sidecar; não há biblioteca embutível documentada.
- **Desempenho publicado:** não encontrado.
- **AuthZEN:** 1.0 (ADS 26.2), endpoints `evaluation`, `evaluations`, descoberta; buscas (subject/resource/action search) não listadas nas release notes.

## Verificação de diferenciais candidatos do projeto
- **Aprovação M de N de primeira classe?** **Não.** Nada na documentação do ADS/Hub/APD. XACML tem "obligations" que poderiam exigir aprovação, mas o fluxo de aprovação teria de ser implementado pelo PEP.
- **Autorização sensível ao estado / transições declarativas?** **Não.** Possível só como atributo (`resource.status`) e ação nomeada; sem definição de estados/transições.

## Implicações para o projeto
1. **Adotar:** obrigações como cidadãs de primeira classe, alinhadas ao perfil de obrigações do AuthZEN (que herda do XACML).
2. **Adotar:** testes de política em níveis (unidade com atributos fixos; integração separada) e execução no CI — mesma recomendação da Axiomatics.
3. **Adotar:** reverse query / avaliação parcial para listagem e filtros de busca em ABAC.
4. **Evitar:** linguagem baseada em XML e modelo de combinação complexo exposto ao usuário; se tivermos linguagem própria, semântica de combinação simples (deny vence) e explícita.
5. **Evitar:** quebrar credenciais em upgrade por troca de dependência; formatos de hash versionados e migração automática.
6. **Oportunidade:** Axiomatics agora é parte de um grupo de defesa; clientes não-defesa e de custo baixo ficam sem opção ABAC madura e aberta.

## Fontes
- https://docs.axiomatics.com/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/platform-overview/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/access-decision-service/ — 2026-11-03 — oficial (versão 26.2.1)
- https://docs.axiomatics.com/access-decision-service/api/ — 2026-11-03 — oficial (OpenAPI com endpoints AuthZEN)
- https://docs.axiomatics.com/access-decision-service/release-notes/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/access-decision-service/concepts/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/access-decision-service/deploy/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/contextual-authorization-query/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/axiomatics-policy-devops/using-apd/testing-alfa-policies/ — 2026-11-03 — oficial
- https://docs.axiomatics.com/authorization-domain-manager/ — 2026-11-03 — oficial
- https://alfa.guide/ — 2026-11-03 — site da linguagem (mantido pela Axiomatics)
- https://www.leonardo.com/en/press-release-detail/-/detail/09-07-2025-leonardo-acquires-axiomatics-of-sweden — 2026-11-03 — press release do comprador
- https://www.defensenews.com/global/europe/2025/09/10/leonardo-snaps-up-nordic-cybersecurity-firms-in-bid-to-shift-focus/ — 2026-11-03 — imprensa
- https://www.ietf.org/archive/id/draft-brossard-oauth-rar-authzen-03.txt — 2026-11-03 — rascunho IETF (expirado)
- https://openid.net/see-authzen-in-action-at-gartner-iam-2025/ — 2026-11-03 — OpenID Foundation
- https://axiomatics.com/events/north-america/join-us-at-identiverse — 2026-11-03 — fornecedor
- https://lists.oasis-open.org/archives/xacml/201603/msg00014.html — 2026-11-03 — lista do TC XACML (opinião de participante)
- https://packetstormsecurity.com/files/cve/CVE-2016-4312 — 2026-11-03 — CVE de terceiro (WSO2) no ecossistema XACML
- https://services.nvd.nist.gov/rest/json/cves/2.0?keywordSearch=Axiomatics — 2026-11-03 — NVD (0 resultados)
