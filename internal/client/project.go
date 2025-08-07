// client.go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)


type CreateProjectRequest struct {
    ProjectTitle   string `json:"projectTitle"`
    OrganizationID string `json:"organizationId"`
    Name           string `json:"name"`
}

type CreateProjectResponse struct {
    ID             string   `json:"id"`
    DateCreated    string   `json:"dateCreated"`
    DateUpdated    string   `json:"dateUpdated"`
    Name           string   `json:"name"`
    Website        string   `json:"website"`
    OrganizationID string   `json:"organizationId"`
    Projects       []Project `json:"projects"`
}

type Project struct {
    ID               string `json:"id"`
    Title            string `json:"title"`
    OwnerID          string `json:"ownerId"`
    TeamID           string `json:"teamId"`
    IsConnectProject bool   `json:"isConnectProject"`
    IsHidden         bool   `json:"isHidden"`
}

type ProjectsResponse struct {
    Items          []Project    `json:"items"`
    NextPageCursor interface{}  `json:"nextPageCursor"`
}

func (c *Client) CreateProject(ctx context.Context, organizationID, projectName string) (*Project, *Project, error) {
    url := fmt.Sprintf("%s/teams?organizationId=%s", c.baseURL, organizationID)

    reqBody := CreateProjectRequest{
        ProjectTitle:   projectName,
        OrganizationID: organizationID,
        Name:           projectName,
    }
    jsonBody, _ := json.Marshal(reqBody)

    req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.accessToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, nil, fmt.Errorf("failed to create project with status code: %d", resp.StatusCode)
    }

    var createProjectResp CreateProjectResponse
    err = json.NewDecoder(resp.Body).Decode(&createProjectResp)
    if err != nil {
        return nil, nil, err
    }

    var connectProject *Project
    var automateProject *Project
    for _, project := range createProjectResp.Projects {
        if project.IsConnectProject {
            tflog.Debug(ctx, fmt.Sprintf("connect project ID: %s", project.ID))
            connectProject = &Project{
                ID:               project.ID,
                Title:            project.Title,
                OwnerID:          project.OwnerID,
                TeamID:           project.TeamID,
                IsConnectProject: project.IsConnectProject,
                IsHidden:         project.IsHidden,
            }
        } else {
            tflog.Debug(ctx, fmt.Sprintf("NON connect project ID (older): %s", project.ID))
            automateProject = &Project{
                ID:               project.ID,
                Title:            project.Title,
                OwnerID:          project.OwnerID,
                TeamID:           project.TeamID,
                IsConnectProject: project.IsConnectProject,
                IsHidden:         project.IsHidden,
            }
        }
    }

    if connectProject == nil {
        return nil, nil, fmt.Errorf("connect project not found in the response")
    }

    if automateProject == nil {
        // That means that the older project is no longer created. We should return the connect project only
        return connectProject, nil, nil
    }

    return connectProject, automateProject, nil
}

func (c *Client) GetProjects(ctx context.Context, teamID string) ([]Project, error) {
    // Use reasonable page size to get projects efficiently
    url := fmt.Sprintf("%s/projects?teamId=%s&size=9007199254740991", c.baseURL, teamID)

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
        return nil, fmt.Errorf("failed to get projects for team_id %s with status code: %d", teamID, resp.StatusCode)
    }

    var projectsResponse ProjectsResponse
    err = json.NewDecoder(resp.Body).Decode(&projectsResponse)
    if err != nil {
        return nil, err
    }

    return projectsResponse.Items, nil
}

func (c *Client) GetProjectByID(ctx context.Context, projectID, teamID string) (*Project, error) {
    url := fmt.Sprintf("%s/projects/%s?teamId=%s", c.baseURL, projectID, teamID)

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
        return nil, fmt.Errorf("failed to get project with status code: %d", resp.StatusCode)
    }

    var project Project
    err = json.NewDecoder(resp.Body).Decode(&project)
    if err != nil {
        return nil, err
    }

    return &project, nil
}

type UpdateProjectTitleRequest struct {
    Title string `json:"title"`
}

func (c *Client) UpdateProjectTitle(ctx context.Context, projectID, teamID, newTitle string) (*Project, error) {
    url := fmt.Sprintf("%s/projects/%s?teamId=%s", c.baseURL, projectID, teamID)

    reqBody := UpdateProjectTitleRequest{
        Title: newTitle,
    }
    jsonBody, _ := json.Marshal(reqBody)

    req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+c.accessToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to update project title with status code: %d", resp.StatusCode)
    }

    var updatedProject Project
    err = json.NewDecoder(resp.Body).Decode(&updatedProject)
    if err != nil {
        return nil, err
    }

    return &updatedProject, nil
}

func (c *Client) DeleteProject(ctx context.Context, projectID, teamID string) error {
    url := fmt.Sprintf("%s/projects/%s?teamId=%s", c.baseURL, projectID, teamID)
    tflog.Debug(ctx, fmt.Sprintf("url to delete: %s", url))
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

    tflog.Debug(ctx, fmt.Sprintf("delete response: %d", resp.StatusCode))
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to delete project with status code: %d", resp.StatusCode)
    }

    return nil
}
