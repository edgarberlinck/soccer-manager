# Training System Design

## Goal

Implementar o sistema de treino do jogo com duracao, intensidade, desgaste fisico e ganho de atributos, levando em conta idade e estado fisico do jogador.

## Scope

Esta iteracao cobre apenas o dominio e a simulacao de treino:

- duracao do treino
- intensidade do treino
- desgaste fisico do jogador
- ganho de atributo principal do treino
- influencia de idade e condicao fisica no ganho
- variacao aleatoria controlada
- finalizacao de treino

Esta iteracao nao cobre:

- persistencia do treino finalizado
- UI de treino
- risco de lesao
- multiplos atributos afetados por uma mesma sessao
- folga/descanso como fluxo completo de negocio

## Domain Model

### Player

`player.Attributes.FisicalStatus` representa a condicao fisica atual do jogador em escala de `0..100`.

Regras:

- `0` representa jogador esgotado
- `100` representa jogador totalmente descansado
- treino reduz `FisicalStatus`
- descanso no futuro recuperara `FisicalStatus`

### Training Type

Cada `TrainingType` define:

- duracao base
- ganho base
- desgaste base
- atributo principal afetado

Nesta iteracao, cada tipo de treino afeta apenas um atributo principal.

### Intensity

As intensidades sao:

- `Soft`
- `Medium`
- `Intense`

Cada intensidade define multiplicadores para:

- duracao
- ganho
- desgaste

Regras de design:

- `Soft` dura menos, desgasta menos e rende menos
- `Medium` e o baseline
- `Intense` dura mais, desgasta mais e rende mais

## Gain Formula

O ganho final do treino deve ser calculado como:

`final gain = base gain x intensity multiplier x age factor x physical factor x random factor`

### Age factor

A idade reduz o ganho de forma leve:

- jogadores jovens ganham um pequeno bonus
- jogadores no auge ficam proximos do baseline
- jogadores mais velhos recebem uma pequena penalidade

A penalidade por idade nao deve inviabilizar evolucao de veteranos.

### Physical factor

A condicao fisica afeta o rendimento de forma mais forte que a idade:

- jogador descansado tem aproveitamento alto
- jogador em condicao media tem leve reducao
- jogador desgastado tem reducao forte

### Random factor

O fator aleatorio existe apenas para variar ligeiramente o resultado.

Regras:

- a variacao deve ser pequena
- a sorte nao deve superar tipo, intensidade, idade e condicao fisica
- dois treinos similares podem render resultados um pouco diferentes

## Training Flow

### StartTraining

Ao iniciar o treino:

- a sessao recebe sua duracao final calculada a partir de tipo + intensidade
- `StartedAt` e definido
- `EndsAt` e definido com base na duracao calculada
- o treino fica com status `In Progress`

### FinishTraining

Ao finalizar o treino:

- o atributo principal recebe o ganho calculado
- `FisicalStatus` e reduzido pelo desgaste calculado
- `FisicalStatus` permanece no intervalo `0..100`
- o treino muda para status `Finished`
- o resumo do treino e preenchido

## Constraints

- `FisicalStatus` nunca pode ficar abaixo de `0`
- `FisicalStatus` nunca pode ficar acima de `100`
- o ganho final nunca pode ser negativo
- todo treino deve produzir algum desgaste minimo
- `Intense` deve tender a render mais que `Medium`
- `Medium` deve tender a render mais que `Soft`
- condicao fisica baixa deve reduzir bem o aproveitamento

## Initial Implementation Boundaries

Arquivos-alvo desta iteracao:

- `internal/domain/training/training.go`
- `simulation/training.go`
- `simulation/training_test.go`
- `engine/engine.go`

## Testing Strategy

Cobrir pelo menos:

- duracao calculada por tipo e intensidade
- `StartTraining` definindo `EndsAt` corretamente
- treino intenso rendendo mais e desgastando mais que treino medio
- treino leve rendendo menos e desgastando menos
- idade mais alta reduzindo ganho sem zerar evolucao
- `FisicalStatus` baixo reduzindo ganho de forma clara
- `FinishTraining` alterando atributo principal e condicao fisica
- limites `0..100` de `FisicalStatus`

## Open Future Work

Fases futuras podem incluir:

- folga/descanso com recuperacao de `FisicalStatus`
- risco de lesao
- diferentes perfis de treino por posicao
- impacto de treinador/estrutura
- persistencia e processamento via scheduler
