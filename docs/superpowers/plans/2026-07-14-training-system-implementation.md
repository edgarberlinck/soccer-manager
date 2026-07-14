# Training System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implementar o sistema de treino com duracao, ganho e desgaste derivados de tipo e intensidade, modulados por idade e condicao fisica.

**Architecture:** As regras de treino ficam centralizadas no dominio `internal/domain/training`, que calcula duracao, desgaste e ganho efetivo. A simulacao passa a iniciar e finalizar treinos usando essas regras, enquanto a engine so orquestra e expõe a interface de alto nivel.

**Tech Stack:** Go, pacote de dominio interno, testes com `go test`.

## Global Constraints

- `player.Attributes.FisicalStatus` representa a condicao fisica atual do jogador na escala `0..100`.
- `Soft` dura menos, desgasta menos e rende menos que `Medium`.
- `Intense` dura mais, desgasta mais e rende mais que `Medium`.
- O ganho final depende de tipo, intensidade, idade, `FisicalStatus` e variacao aleatoria controlada.
- O ganho final nunca pode ser negativo.
- `FisicalStatus` nunca pode sair do intervalo `0..100`.
- Cada treino afeta apenas um atributo principal nesta iteracao.
- Esta iteracao nao cobre persistencia, UI, risco de lesao ou fluxo completo de folga.

---

### Task 1: Formalizar regras do dominio de treino

**Files:**

- Modify: `internal/domain/training/training.go`
- Test: `internal/domain/training/training_test.go`

**Interfaces:**

- Produces: `func (t TrainingType) BaseDuration() time.Duration`
- Produces: `func (t TrainingType) BaseGain() float64`
- Produces: `func (t TrainingType) BaseFatigueCost() int`
- Produces: `func (t TrainingType) TargetAttribute(current player.Attributes) (getter func() int, setter func(int) player.Attributes)` or equivalent helper functions anchored in the package
- Produces: `func (i Intensity) DurationMultiplier() float64`
- Produces: `func (i Intensity) GainMultiplier() float64`
- Produces: `func (i Intensity) FatigueMultiplier() float64`
- Produces: `func (s TrainingSession) ResolvedDuration() time.Duration`
- Produces: `func ComputeTrainingGain(session TrainingSession, athlete player.Player, randomFactor float64) int`
- Produces: `func ComputeFatigueCost(session TrainingSession) int`

- [ ] **Step 1: Write failing domain tests**

```go
package training

import (
    "manager/game/internal/domain/player"
    "testing"
    "time"
)

func TestResolvedDurationVariesByIntensity(t *testing.T) {
    base := TrainingSession{Type: Passing, Intensity: Medium}.ResolvedDuration()
    light := TrainingSession{Type: Passing, Intensity: Soft}.ResolvedDuration()
    intense := TrainingSession{Type: Passing, Intensity: Intense}.ResolvedDuration()

    if !(light < base && base < intense) {
        t.Fatalf("expected Soft < Medium < Intense durations, got %s %s %s", light, base, intense)
    }
}

func TestComputeTrainingGainFallsWithAge(t *testing.T) {
    session := TrainingSession{Type: Passing, Intensity: Medium}
    younger := player.Player{Age: 20, Attributes: player.Attributes{FisicalStatus: 90}}
    older := player.Player{Age: 34, Attributes: player.Attributes{FisicalStatus: 90}}

    youngGain := ComputeTrainingGain(session, younger, 1.0)
    oldGain := ComputeTrainingGain(session, older, 1.0)

    if youngGain <= oldGain {
        t.Fatalf("expected younger player to gain more, got young=%d old=%d", youngGain, oldGain)
    }
}

func TestComputeTrainingGainFallsWithPhysicalStatus(t *testing.T) {
    session := TrainingSession{Type: Passing, Intensity: Medium}
    fit := player.Player{Age: 24, Attributes: player.Attributes{FisicalStatus: 90}}
    tired := player.Player{Age: 24, Attributes: player.Attributes{FisicalStatus: 35}}

    fitGain := ComputeTrainingGain(session, fit, 1.0)
    tiredGain := ComputeTrainingGain(session, tired, 1.0)

    if fitGain <= tiredGain {
        t.Fatalf("expected fitter player to gain more, got fit=%d tired=%d", fitGain, tiredGain)
    }
}

func TestComputeFatigueCostVariesByIntensity(t *testing.T) {
    light := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Soft})
    medium := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Medium})
    intense := ComputeFatigueCost(TrainingSession{Type: Passing, Intensity: Intense})

    if !(light < medium && medium < intense) {
        t.Fatalf("expected Soft < Medium < Intense fatigue, got %d %d %d", light, medium, intense)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/training`
Expected: FAIL with missing methods/functions such as `ResolvedDuration`, `ComputeTrainingGain`, or `ComputeFatigueCost`

- [ ] **Step 3: Write minimal domain implementation**

Implement in `internal/domain/training/training.go`:

```go
func (t TrainingType) BaseDuration() time.Duration { /* fixed per type */ }
func (t TrainingType) BaseGain() float64 { /* fixed per type */ }
func (t TrainingType) BaseFatigueCost() int { /* fixed per type */ }
func (i Intensity) DurationMultiplier() float64 { /* Soft < Medium < Intense */ }
func (i Intensity) GainMultiplier() float64 { /* Soft < Medium < Intense */ }
func (i Intensity) FatigueMultiplier() float64 { /* Soft < Medium < Intense */ }
func (s TrainingSession) ResolvedDuration() time.Duration { /* base * multiplier */ }
func ComputeTrainingGain(session TrainingSession, athlete player.Player, randomFactor float64) int { /* clamp >= 0, age and fitness aware */ }
func ComputeFatigueCost(session TrainingSession) int { /* clamp >= 1 */ }
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/training`
Expected: PASS

### Task 2: Aplicar regras na simulacao de treino

**Files:**

- Modify: `simulation/training.go`
- Modify: `simulation/training_test.go`
- Test: `simulation/training_test.go`

**Interfaces:**

- Consumes: `func (s TrainingSession) ResolvedDuration() time.Duration`
- Consumes: `func ComputeTrainingGain(session TrainingSession, athlete player.Player, randomFactor float64) int`
- Consumes: `func ComputeFatigueCost(session TrainingSession) int`
- Produces: `func StartTraining(player player.Player, session training.TrainingSession) training.Training`
- Produces: `func FinishTraining(session training.Training, now time.Time, randomFactor float64) training.Training`

- [ ] **Step 1: Write failing simulation tests**

Add tests covering:

```go
func TestStartTrainingUsesResolvedDuration(t *testing.T) {}
func TestFinishTrainingMarksSessionFinishedAndAppliesGain(t *testing.T) {}
func TestFinishTrainingReducesPhysicalStatusWithoutLeavingBounds(t *testing.T) {}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./simulation`
Expected: FAIL with missing `FinishTraining` or wrong duration/status behavior

- [ ] **Step 3: Write minimal simulation implementation**

Implement in `simulation/training.go`:

```go
func StartTraining(player player.Player, session training.TrainingSession) training.Training {
    session.Duration = session.ResolvedDuration()
    startedAt := time.Now()
    return training.Training{ /* fill StartedAt, EndsAt, Status */ }
}

func FinishTraining(current training.Training, now time.Time, randomFactor float64) training.Training {
    // compute gain and fatigue, apply target attribute update and clamp FisicalStatus
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./simulation`
Expected: PASS

### Task 3: Concluir engine de treino

**Files:**

- Modify: `engine/engine.go`
- Test: `engine/engine_test.go`

**Interfaces:**

- Consumes: `func StartTraining(player player.Player, session training.TrainingSession) training.Training`
- Consumes: `func FinishTraining(session training.Training, now time.Time, randomFactor float64) training.Training`
- Produces: `func (e *Engine) StartTraining(player player.Player, session training.TrainingSession) training.Training`
- Produces: `func (e *Engine) FinishTraining(session training.Training) training.Training`

- [ ] **Step 1: Write failing engine tests**

Add tests covering:

```go
func TestEngineStartTrainingReturnsTraining(t *testing.T) {}
func TestEngineFinishTrainingReturnsFinishedTraining(t *testing.T) {}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./engine`
Expected: FAIL because engine methods do not return values or do not finalize training

- [ ] **Step 3: Write minimal engine implementation**

Update `engine/engine.go`:

```go
func (e *Engine) StartTraining(player player.Player, session training.TrainingSession) training.Training {
    return simulation.StartTraining(player, session)
}

func (e *Engine) FinishTraining(session training.Training) training.Training {
    return simulation.FinishTraining(session, time.Now(), 1.0)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./engine`
Expected: PASS

### Task 4: Regressao e integracao local

**Files:**

- Modify: none
- Test: existing packages

**Interfaces:**

- Consumes: all interfaces produced above

- [ ] **Step 1: Run focused package tests**

Run: `go test ./internal/domain/training ./simulation ./engine`
Expected: PASS

- [ ] **Step 2: Run broader compilation check**

Run: `go test ./...`
Expected: PASS
