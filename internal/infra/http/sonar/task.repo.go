package sonar

import "sonarbridge-go/configs"

type TaskRepository struct {
	Config *configs.Config
}

func NewSonarTaskRepo() *TaskRepository {
	return &TaskRepository{
		Config: nil,
	}
}

func (t *TaskRepository) GetTaskById(id string) (*TaskResponse, error) {
	// client, err := t.GetTaskById()
	return nil, nil
}
