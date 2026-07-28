# 🤖 Sistema de Bots - Resumo Executivo

## O que foi criado?

Sistema completo de **jogadores controlados por IA** que são indistinguíveis de jogadores humanos. Os 2000 bots garantem que sempre há oponentes disponíveis para jogar, mesmo quando não há outros jogadores online.

**Conceito Chave:** Bots são adversários de IA que aparecem como jogadores normais em todos os lugares - perfis, rankings, histórico de partidas. Nenhum jogador sabe que está jogando contra IA.

## ✅ Características

- ✅ 2000 jogadores IA indistinguíveis de humanos
- ✅ Cada bot tem seu próprio time com 15 jogadores
- ✅ Estratégias variadas (Conservador, Equilibrado, Agressivo)
- ✅ Aceitam partidas automaticamente
- ✅ Escalam times inteligentemente
- ✅ Treinam jogadores nos bastidores
- ✅ Sem indicação visual de que são bots

## 📁 Arquivos Criados

### Database
- `api/internal/infrastructure/database/migrations/00008_add_bot_support.sql` - Migration para suporte a bots
- `api/internal/infrastructure/database/queries/bot.sql` - Queries SQL para bots
- `api/internal/infrastructure/database/queries/player.sql` - Atualizado com queries adicionais

### Domain
- `api/internal/domain/bot/bot.go` - Lógica de negócio dos bots
- `api/internal/domain/bot/bot_test.go` - Testes unitários

### Services
- `api/internal/services/bot_service.go` - Serviço que orquestra ações dos bots

### Commands
- `api/cmd/create-bots/main.go` - Comando para criar os 2000 bots
- `api/cmd/bot-scheduler/main.go` - Scheduler para executar ações periódicas

### Documentação
- `Bots.md` - Documentação completa do sistema
- `BotIntegration.md` - Guia de integração com o jogo
- `setup-bots.sh` - Script de setup automatizado

## 🚀 Como Usar

### 1. Setup Inicial

```bash
# Instalar ferramentas (se necessário)
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Ou usar o script automatizado
chmod +x setup-bots.sh
./setup-bots.sh
```

### 2. Rodar Migration

```bash
cd api
goose -dir internal/infrastructure/database/migrations postgres "$DATABASE_URL" up
```

### 3. Gerar Código (sqlc)

```bash
cd api
sqlc generate
```

### 4. Criar os Bots

```bash
cd api
go run cmd/create-bots/main.go
```

Isso criará:
- 2000 usuários bot (indistinguíveis de humanos)
- 2000 clubes (um para cada bot)
- 30.000 jogadores (15 por clube)
- ~667 Conservadores
- ~667 Equilibrados  
- ~666 Agressivos

### 5. Iniciar Scheduler

```bash
cd api
go run cmd/bot-scheduler/main.go
```

O scheduler executa ações dos bots a cada 30 minutos.

## 🎯 Funcionalidades

### Estratégias

**Conservador:**
- Treinos mais leves
- Descansa quando status físico < 60
- Desenvolvimento lento mas seguro

**Equilibrado:**
- Mix balanceado de treinos
- Descansa quando status físico < 40
- Abordagem moderada

**Agressivo:**
- Treinos intensos
- Descansa apenas quando status físico < 25
- Desenvolvimento rápido, alto desgaste

### Ações Automáticas

- **Treinos:** 1-5 jogadores aleatórios por rodada
- **Tipos:** Finishing, Passing, Dribbling, Shooting, Speed, Strength, Stamina, Goalkeeping
- **Seleção:** Bots sempre escolhem os melhores jogadores
- **Descanso:** Automático baseado na estratégia

## 🧪 Testes

```bash
cd api/internal/domain/bot
go test -v
```

## 📊 Monitoramento

O sistema gera logs detalhados:
```
Executando ações para 2000 bots...
Progresso: 500 bots criados...
✅ Processo concluído!
Bots criados com sucesso: 2000
```

## 🔧 Configurações

### Intervalo do Scheduler
Edite `cmd/bot-scheduler/main.go`:
```go
botService.StartBotScheduler(ctx, 30*time.Minute) // Altere aqui
```

### Jogadores por Rodada
Edite `internal/services/bot_service.go`:
```go
if numToTrain > 5 { // Altere este número
    numToTrain = 5
}
```

## 🔌 Integração

Consulte `BotIntegration.md` para opções de integração:
- Processo separado (recomendado para MVP)
- Integrado no scheduler principal
- Como job no cron

## ⚡ Performance

- Bots não afetam performance da API
- Ações executadas em background
- Pool de workers pode ser implementado se necessário
- Queries otimizadas com índices

## 🎮 Gameplay

### Como os bots interagem:

1. **Treinos Automáticos**
   - Executam a cada rodada do scheduler
   - Escolhem tipo e intensidade baseado na estratégia
   - Aplicam ganhos e fadiga automaticamente

2. **Seleção de Time**
   - Sempre usam os melhores jogadores
   - Consideram overall (média de atributos)
   - Podem ser expandidos para considerar posições

3. **Gestão de Fadiga**
   - Conservadores: muito cautelosos
   - Equilibrados: moderados
   - Agressivos: arriscam mais

## 🚧 Futuras Melhorias

- [ ] Partidas automáticas entre bots
- [ ] Transferências inteligentes
- [ ] Ligas exclusivas para bots
- [ ] Sistema de reputação
- [ ] Eventos aleatórios (lesões, talentos descobertos)
- [ ] Táticas e formações variadas
- [ ] Gestão financeira automática

## 📝 Notas Técnicas

- Bots usam `is_bot = true` na tabela users
- Password hash vazio (bots não fazem login)
- Todos ativos por padrão (`active = true`)
- Estratégia armazenada no campo `bot_strategy`
- Queries otimizadas com índices

## ✅ Checklist de Deploy

- [ ] Rodar migration 00008
- [ ] Gerar código sqlc
- [ ] Testar criação de bots
- [ ] Validar queries funcionando
- [ ] Iniciar scheduler
- [ ] Monitorar logs
- [ ] Verificar performance

## 🤝 Contribuindo

Para adicionar novas estratégias:

1. Adicione em `bot.go`:
```go
const (
    Conservador Strategy = "conservador"
    Equilibrado Strategy = "equilibrado"
    Agressivo   Strategy = "agressivo"
    MinhaEstrategia Strategy = "minha_estrategia" // NOVO
)
```

2. Implemente lógica em `selectIntensity()`

3. Atualize `create-bots/main.go` para incluir nova estratégia

4. Adicione testes em `bot_test.go`

---

**Status:** ✅ Pronto para uso
**Versão:** 1.0
**Última atualização:** 2024-07-27
