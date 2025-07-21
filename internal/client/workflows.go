package client

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type WorkflowStep struct {
    ID           string                 `json:"id"`
    DateCreated  string                 `json:"dateCreated"`
    DateUpdated  string                 `json:"dateUpdated"`
    Description  string                 `json:"description"`
    Type         string                 `json:"type"`
    Parameters   map[string]interface{} `json:"parameters"`
    Next         string                 `json:"next,omitempty"`
    Status       string                 `json:"status"`
    WorkflowID   string                 `json:"workflowId"`
    Migrations   map[string]interface{} `json:"_migrations"`
}

type Workflow struct {
    ID                    string         `json:"id"`
    DateCreated          string         `json:"dateCreated"`
    DateUpdated          string         `json:"dateUpdated"`
    Description          string         `json:"description"`
    ProjectID            string         `json:"projectId"`
    TeamID               string         `json:"teamId"`
    IntegrationID        string         `json:"integrationId"`
    WorkflowVersion      int            `json:"workflowVersion"`
    Tags                 []string       `json:"tags"`
    IsOnboardingWorkflow bool           `json:"isOnboardingWorkflow"`
    Steps                []WorkflowStep `json:"steps,omitempty"`
}

// WorkflowsResponse represents the paginated response from the workflows API
type WorkflowsResponse struct {
    Items          []Workflow  `json:"items"`
    NextPageCursor *string     `json:"nextPageCursor"`
}

func (c *Client) GetWorkflows(ctx context.Context, projectID, integrationID string) ([]Workflow, error) {
    url := fmt.Sprintf("%s/projects/%s/workflows?includeDeleted=false&integrationId=%s", c.baseURL, projectID, integrationID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer "+c.accessToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get workflows with status code: %d", resp.StatusCode)
    }

    var workflowsResponse WorkflowsResponse
    err = json.NewDecoder(resp.Body).Decode(&workflowsResponse)
    if err != nil {
        return nil, err
    }

    return workflowsResponse.Items, nil
}