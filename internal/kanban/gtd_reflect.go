package kanban

type ReflectItem struct {
	ItemID                  string
	Class                   GTDClass
	DependsOn               []string
	Completed               bool
	Cancelled               bool
	EvidenceRevision        int64
	CurrentEvidenceRevision int64
}

type ReflectInput struct {
	Items     []ReflectItem
	Relations []GTDRelation
}

type ReflectResult struct {
	Blocked         map[string]bool
	Stale           map[string]bool
	NeedsNextAction map[string]bool
}

func ReflectGTDState(in ReflectInput) ReflectResult {
	result := ReflectResult{Blocked: map[string]bool{}, Stale: map[string]bool{}, NeedsNextAction: map[string]bool{}}
	byID := make(map[string]ReflectItem, len(in.Items))
	projectHasAction := map[string]bool{}
	for _, item := range in.Items {
		byID[item.ItemID] = item
		if item.EvidenceRevision != item.CurrentEvidenceRevision {
			result.Stale[item.ItemID] = true
		}
	}
	for _, relation := range in.Relations {
		if relation.Kind != RelationPartOf {
			continue
		}
		action, actionOK := byID[relation.SubjectID]
		project, projectOK := byID[relation.ObjectID]
		if actionOK && projectOK && action.Class == ClassAction && project.Class == ClassProject && !action.Completed && !action.Cancelled {
			projectHasAction[project.ItemID] = true
		}
	}
	for _, item := range in.Items {
		for _, depID := range item.DependsOn {
			dep, ok := byID[depID]
			if !ok || !dep.Completed || dep.Cancelled {
				result.Blocked[item.ItemID] = true
			}
		}
		if item.Class == ClassProject && !projectHasAction[item.ItemID] {
			result.NeedsNextAction[item.ItemID] = true
		}
	}
	return result
}
