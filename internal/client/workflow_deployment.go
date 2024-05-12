
package client

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type WorkflowDeployment struct {
    ID       string `json:"id"`
    Status   string `json:"status"`
    IsActive bool   `json:"isActive"`
}

type WorkflowMigration struct {
    ID         string              `json:"id"`
    WorkflowID string              `json:"workflowId"`
    Deployment WorkflowDeployment `json:"deployment"`
}


    func (c *Client) CreateWorkflowDeployment(ctx context.Context, projectID, workflowID string) (string, error) {
        url := fmt.Sprintf("%s/projects/%s/workflows/%s/deployments", c.baseURL, projectID, workflowID)

        req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
        if err != nil {
            return "", err
        }
        req.Header.Set("Authorization", "Bearer "+c.accessToken)

        resp, err := c.httpClient.Do(req)
        if err != nil {
            return "", err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusCreated {
   	        errorMessage := formatErrorMessage(resp)
   	        return "", fmt.Errorf("%s, status code: %d", errorMessage, resp.StatusCode)
        }

        var response struct {
            ID     string `json:"id"`
            Status string `json:"status"`
        }
        err = json.NewDecoder(resp.Body).Decode(&response)
        if err != nil {
            return "", err
        }

        return response.ID, nil
    }

func (c *Client) GetWorkflowDeployment(ctx context.Context, projectID, deploymentID string) (*WorkflowDeployment, error) {
    url := fmt.Sprintf("%s/projects/%s/deployments/%s", c.baseURL, projectID, deploymentID)

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
        return nil, fmt.Errorf("failed to get workflow deployment with status code: %d", resp.StatusCode)
    }

    var deployment WorkflowDeployment
    err = json.NewDecoder(resp.Body).Decode(&deployment)
    if err != nil {
        return nil, err
    }

    return &deployment, nil
}

func (c *Client) DeleteWorkflowDeployment(ctx context.Context, projectID, workflowID string) error {
    url := fmt.Sprintf("%s/projects/%s/workflows/%s/deployments", c.baseURL, projectID, workflowID)

    req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Bearer "+c.accessToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return nil
    }

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to delete workflow deployment with status code: %d", resp.StatusCode)
    }

    return nil
}

func (c *Client) GetLatestWorkflowMigration(ctx context.Context, projectID, workflowID string) (*WorkflowMigration, error) {
    url := fmt.Sprintf("%s/projects/%s/workflows/migrations/latest", c.baseURL, projectID)

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
        return nil, fmt.Errorf("failed to get latest workflow migrations with status code: %d", resp.StatusCode)
    }

    var migrations []WorkflowMigration
    err = json.NewDecoder(resp.Body).Decode(&migrations)
    if err != nil {
        return nil, err
    }

    for _, migration := range migrations {
        if migration.WorkflowID == workflowID {
            return &migration, nil
        }
    }

    return nil, nil
}