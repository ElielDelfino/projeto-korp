# Projeto Korp — Serviço HTTP em Go, Observabilidade e Automação

Serviço HTTP em **Go** exposto através de um **proxy reverso Nginx**, instrumentado com
métricas no padrão **Prometheus**, visualizado no **Grafana** e provisionado de ponta a
ponta com **Ansible**. Desenvolvido como desafio técnico de estágio DevOps.

---

## Arquitetura

```
   curl                       rede bridge "korp-net"
  (cliente)                   (DNS interno por nome de serviço)
     │
     │ :80
     ▼
┌──────────┐   proxy_pass     ┌──────────────────────────┐
│  nginx   │ ───────────────► │ http-server-projeto-korp │
│  :80     │   :8080          │  :8080 (NÃO publicado)   │
└──────────┘                  └────────────┬─────────────┘
 publicado                                 │ /metrics (scrape)
 no host                                   ▼
                              ┌────────────┐   query    ┌──────────┐
                              │ prometheus │ ◄────────── │ grafana  │
                              │  :9090     │            │  :3000   │
                              └────────────┘            └──────────┘
```

- Só o **Nginx** (porta 80) e as UIs de **Prometheus** (9090) e **Grafana** (3000) são
  publicados no host.
- O **serviço Go não publica porta**: é alcançado apenas pela rede interna, por DNS
  (`http-server-projeto-korp:8080`).

---

## Stack

| Camada          | Tecnologia                              |
|-----------------|-----------------------------------------|
| Serviço         | Go 1.26 (`net/http` + `prometheus/client_golang`) |
| Containerização | Docker (multi-stage build)              |
| Orquestração    | Docker Compose                          |
| Proxy reverso   | Nginx (imagem oficial)                  |
| Observabilidade | Prometheus + Grafana                    |
| Automação       | Ansible                                 |

---

## Estrutura do repositório

```
projeto-korp/
├── app/                              # serviço Go
│   ├── main.go                       # servidor HTTP + endpoint + métricas
│   ├── go.mod / go.sum               # módulo e lockfile (versionados)
│   └── Dockerfile                    # build multi-stage
├── nginx/conf.d/
│   └── http-server-projeto-korp.conf # proxy reverso (uso standalone)
├── prometheus/
│   └── prometheus.yml                # config de scrape (uso standalone)
├── grafana/provisioning/             # datasource + dashboard
│   ├── datasources/datasources.yml
│   └── dashboards/dashboards.yml + http-server-projeto-korp-dashboard.json
├── ansible/                          # automação
│   ├── playbook.yml                  # orquestrador (import_tasks)
│   ├── ansible.cfg / inventory.ini / requirements.yml
│   ├── group_vars/
│   │   └── all.yml                   # variáveis (Prometheus, Grafana, etc)
│   ├── tasks/                        # docker-install, deploy-files, build-and-run, validate
│   └── templates/                    # nginx-proxy.conf.j2, prometheus.yml.j2, .env-*.j2
├── docker-compose.yml
├── .env-prometheus.example           # template Prometheus (versionado)
├── .env-prometheus                   # variáveis Prometheus (dev, ignorado no git)
├── .env-grafana.example              # template Grafana (versionado)
├── .env-grafana                      # variáveis Grafana (dev, ignorado no git)
├── .gitignore
└── README.md
```

---

## Configuração de variáveis de ambiente

As credenciais do **Prometheus** e **Grafana** são gerenciadas através de arquivos `.env`:

### Desenvolvimento (Docker Compose local)

Os arquivos `.env-*` são ignorados no git por segurança. Use os arquivos `.example` como template:

```bash
# Copiar arquivos de exemplo
cp .env-prometheus.example .env-prometheus
cp .env-grafana.example .env-grafana

# Editar com suas credenciais (opcional em dev)
vim .env-prometheus .env-grafana

# Subir os containers
docker compose up -d
```

- **`.env-prometheus.example`** — template do Prometheus (versionado)
- **`.env-prometheus`** — credenciais reais (ignorado via `*.env` no `.gitignore`)
- **`.env-grafana.example`** — template do Grafana (versionado)
- **`.env-grafana`** — credenciais reais (ignorado via `*.env` no `.gitignore`)

### Produção (Ansible)

O **Ansible** gera os arquivos `.env` a partir de templates e variáveis:

- **Templates**: `ansible/templates/.env-*.j2`
- **Variáveis**: `ansible/group_vars/all.yml` (desenvolvimento) ou Ansible Vault (produção)

O playbook substitui as variáveis Jinja2 e cria os arquivos `.env` automaticamente:

```bash
cd ansible

# Editar variáveis antes de rodar
vim group_vars/all.yml

# Provisionar (gera .env-prometheus e .env-grafana)
ansible-playbook playbook.yml -K
```

Para **produção segura**, use Ansible Vault:

```bash
# Criar arquivo encriptado com credenciais sensíveis
ansible-vault create group_vars/secrets.yml

# Rodar playbook pedindo a senha do vault
ansible-playbook playbook.yml -K --ask-vault-pass
```

---

## Pré-requisitos

- Linux (Ubuntu/Debian).
- **Ansible** instalado no controlador.
- A coleção **community.docker** (instalada abaixo).
- O Docker **não** precisa estar instalado — o playbook o instala.

---

## Como executar

### Opção 1 — Ansible (recomendado: provisiona tudo com um comando)

```bash
cd ansible

# Instalar a coleção (uma única vez)
ansible-galaxy collection install -r requirements.yml

# Provisionar TODO o ambiente
ansible-playbook playbook.yml -K        # -K pede a senha do sudo
```

O playbook: instala o Docker, cria a árvore de deploy em `/opt/projeto-korp`, copia os
arquivos, builda a imagem, cria a rede, gera as configs (Nginx e Prometheus), sobe os
containers via Compose e valida o serviço por HTTP, exibindo a resposta no console.

### Opção 2 — Docker Compose (teste manual, sem Ansible)

A rede é declarada como `external` no Compose (o Ansible é o dono dela), então **crie a
rede e os arquivos `.env` antes** de subir manualmente:

```bash
# Criar rede
docker network create korp-network

# Copiar templates de exemplo
cp .env-prometheus.example .env-prometheus
cp .env-grafana.example .env-grafana

# Subir os containers (use as credenciais padrão ou edite conforme necessário)
docker compose up -d --build
```

Se precisar customizar as credenciais:

```bash
vim .env-prometheus .env-grafana
docker compose restart prometheus grafana
```

---

## Validação

```bash
# Resposta esperada (via Nginx): JSON com o horário UTC dinâmico
curl http://localhost/projeto-korp
# {"nome":"Projeto Korp","horario":"17:39:48 UTC"}

# Prova de que o serviço Go NÃO está exposto ao host (deve falhar):
curl http://localhost:8080/projeto-korp
# connection refused
```

---

## Observabilidade

- O serviço expõe `/metrics` no padrão Prometheus, incluindo:
  - **`http_requests_total`** (`counter`) — volume de requisições, com labels `method` e `code`.
  - métricas de runtime do Go e do processo (de série).
- **Prometheus** (`http://localhost:9090`) raspa o serviço a cada 15s.
  Veja **Status → Targets** para o alvo `UP`.
  - Credenciais carregadas de `.env-prometheus` (Docker Compose) ou geradas pelo Ansible.
- **Grafana** (`http://localhost:3000`) já sobe com o datasource e o dashboard **provisionados**.
  - Login automático com credenciais do `.env-grafana` (Docker Compose) ou Ansible.
  - Credenciais padrão: `admin`/`admin` (editáveis em `.env-grafana` ou `group_vars/all.yml`).
  - O dashboard mostra:
    - **Disponibilidade** — `up{job="http-server-projeto-korp"}` (scrape) e taxa de sucesso (2xx).
    - **Volume de requisições** — `rate(http_requests_total[...])`.

---

## Endpoints e portas

| Serviço      | URL                                   | Porta no host |
|--------------|---------------------------------------|---------------|
| Nginx (entrada) | http://localhost/projeto-korp      | 80            |
| Prometheus   | http://localhost:9090                 | 9090          |
| Grafana      | http://localhost:3000                 | 3000          |
| Serviço Go   | *interno (não publicado)*             | —             |

---

## Decisões técnicas

- **Multi-stage build** + binário estático (`CGO_ENABLED=0`): imagem final mínima (alpine).
- **Serviço Go sem porta publicada**: única entrada é o Nginx; o backend só é alcançável na
  rede interna. (`EXPOSE 8080` é apenas documentação, não publica porta.)
- **Rede bridge user-defined**: provê DNS por nome de serviço — o Nginx encontra o backend
  por `http-server-projeto-korp:8080`, não por IP.
- **Rede `external`**: o Ansible (`docker_network`) é o dono da rede; o Compose apenas a
  consome — evita conflito de ownership e atende o requisito "criação da rede pelo Ansible".
- **Métricas**: `counter` para volume (grandeza acumulativa); disponibilidade medida pela
  métrica `up` (scrape, vista de fora) e pela taxa de sucesso — não por um gauge tautológico.
- **Variáveis de ambiente organizadas**:
  - Docker Compose usa `env_file: - .env-*` para carregar credenciais por serviço.
  - Ansible gera `.env` via templates Jinja2, permitindo parametrização por ambiente.
  - Credenciais são ignoradas no git (`*.env` no `.gitignore`) e protegidas com modo `0600`.
- **Ansible idempotente e modular**: tasks separadas por responsabilidade (`import_tasks`),
  configs parametrizadas por `template`, e re-execução converge ao mesmo estado.