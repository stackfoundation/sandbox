package image

import (
	"fmt"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func collectCherryPicks(coordinator coordinator.Coordinator, sc *context.StepContext, step *v1.WorkflowStep) {
	cherryPicks := step.CherryPick()
	picks := make(map[string]*v1.Pick, len(cherryPicks))

	if cherryPicks != nil {
		for _, cherryPick := range cherryPicks {
			var pick *v1.Pick
			p, ok := picks[cherryPick.Step]
			if !ok {
				if step.Name() != cherryPick.Step {
					pickSourceImage, err := commitStepImage(coordinator, sc, cherryPick.Step)
					if err != nil {
						fmt.Printf("Error while cherry-picking file from step %v: %v\n", cherryPick.Step, err.Error())
					} else {
						pick = &v1.Pick{
							GeneratedBaseImage: pickSourceImage,
						}
						picks[cherryPick.Step] = pick
					}
				} else {
					pick = &v1.Pick{
						GeneratedBaseImage: "",
					}
					picks[cherryPick.Step] = pick
				}
			} else {
				pick = p
			}

			if pick != nil {
				pick.Copies = append(pick.Copies, cherryPick.From+" "+cherryPick.To)
			}
		}

		if len(picks) > 0 {
			for _, pick := range picks {
				step.State.Picks = append(step.State.Picks, *pick)
			}
		}
	}
}
