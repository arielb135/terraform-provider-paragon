// client.go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type Integration struct {
    ID                  string             `json:"id"`
    DateCreated         string             `json:"dateCreated"`
    DateUpdated         string             `json:"dateUpdated"`
    ProjectID           string             `json:"projectId"`
    CustomIntegrationID string             `json:"customIntegrationId"`
    Type                string             `json:"type"`
    IsActive            bool               `json:"isActive"`
    CustomIntegration   *CustomIntegration `json:"customIntegration"`
    ConnectedUserCount  int                `json:"connectedUserCount"`
}

type IntegrationConfig struct {
    ID           string                 `json:"id"`
    DateCreated  string                 `json:"dateCreated"`
    DateUpdated  string                 `json:"dateUpdated"`
    IntegrationID string                `json:"integrationId"`
}

type CustomIntegration struct {
    ID                 string      `json:"id"`
    DateCreated        string      `json:"dateCreated"`
    DateUpdated        string      `json:"dateUpdated"`
    ProjectID          string      `json:"projectId"`
    Name               string      `json:"name"`
    AuthenticationType string      `json:"authenticationType"`
    Slug               string      `json:"slug"`
}

func (c *Client) GetIntegrations(ctx context.Context, projectID string) ([]Integration, error) {
    url := fmt.Sprintf("%s/projects/%s/integrations", c.baseURL, projectID)

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
        return nil, fmt.Errorf("failed to get integrations with status code: %d", resp.StatusCode)
    }

    var integrations []Integration
    err = json.NewDecoder(resp.Body).Decode(&integrations)
    if err != nil {
        return nil, err
    }

    return integrations, nil
}

type Credential struct {
    ID            string `json:"id"`
    DateCreated   string `json:"dateCreated"`
    DateUpdated   string `json:"dateUpdated"`
    Name          string `json:"name"`
    ProjectID     string `json:"projectId"`
    IntegrationID string `json:"integrationId"`
    Provider      string `json:"provider"`
    Scheme        string `json:"scheme"`
    OnboardingOnly bool   `json:"onboardingOnly"`
    Status         string `json:"status"`
    DateRefreshed  string `json:"dateRefreshed"`
    DateValidUntil string `json:"dateValidUntil"`
}

func (c *Client) GetCredentials(ctx context.Context, projectID string) ([]Credential, error) {
    url := fmt.Sprintf("%s/projects/%s/credentials", c.baseURL, projectID)

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
        return nil, fmt.Errorf("failed to get credentials with status code: %d", resp.StatusCode)
    }

    var credentials []Credential
    err = json.NewDecoder(resp.Body).Decode(&credentials)
    if err != nil {
        return nil, err
    }

    return credentials, nil
}

func (c *Client) GetUserEmailFromToken() (string, error) {
    // Extract the user email from the access token (c.accessToken)
    // Implement the logic to parse the JWT and retrieve the "email" field
    // Return the email and any error encountered
    // Example implementation:
    // token, _ := jwt.Parse(c.accessToken, nil)
    // claims, _ := token.Claims.(jwt.MapClaims)
    // email, _ := claims["email"].(string)
    // return email, nil

    // Placeholder implementation
    return "user@example.com", nil
}

type CreateIntegrationCredentialsRequest struct {
    Name          string         `json:"name"`
    Values        map[string]any `json:"values"`
    Provider      string         `json:"provider"`
    Scheme        string         `json:"scheme"`
    IntegrationID string         `json:"integrationId"`
}

type OAuthValues struct {
    ClientID     string   `json:"clientId"`
    ClientSecret string   `json:"clientSecret"`
    Scopes       string   `json:"scopes"` // Should be with spaces
}

func (c *Client) CreateIntegrationCredentials(ctx context.Context, projectID string, req CreateIntegrationCredentialsRequest) (*Credential, error) {
    url := fmt.Sprintf("%s/projects/%s/credentials/oauth", c.baseURL, projectID)

    jsonBody, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }

    httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, err
    }
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to create/update integration credentials with status code: %d", resp.StatusCode)
    }

    var credential Credential
    err = json.NewDecoder(resp.Body).Decode(&credential)
    if err != nil {
        return nil, err
    }

    return &credential, nil
}

type DecryptedCredential struct {
    ID            string                 `json:"id"`
    DateCreated   string                 `json:"dateCreated"`
    DateUpdated   string                 `json:"dateUpdated"`
    ProjectID     string                 `json:"projectId"`
    Values        map[string]interface{} `json:"values"`
    Provider      string                 `json:"provider"`
    IntegrationID string                 `json:"integrationId"`
    Scheme        string                 `json:"scheme"`
    Status        string                 `json:"status"`
}

func (c *Client) GetDecryptedCredential(ctx context.Context, projectID, credID string) (*DecryptedCredential, error) {
    url := fmt.Sprintf("%s/projects/%s/credentials/%s/decrypted", c.baseURL, projectID, credID)

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

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("credential not found with status code: %d", resp.StatusCode)
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get decrypted credential with status code: %d", resp.StatusCode)
    }

    var credential DecryptedCredential
    err = json.NewDecoder(resp.Body).Decode(&credential)
    if err != nil {
        return nil, err
    }

    return &credential, nil
}

func (c *Client) DeleteCredentials(ctx context.Context, projectID, credentialsID string) error {
    url := fmt.Sprintf("%s/projects/%s/credentials/%s", c.baseURL, projectID, credentialsID)

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
        return fmt.Errorf("failed to delete credentials with status code: %d", resp.StatusCode)
    }

    return nil
}

func (c *Client) UpdateIntegrationStatus(ctx context.Context, projectID, integrationID string, active bool) (*Integration, error) {
    url := fmt.Sprintf("%s/projects/%s/integrations/%s", c.baseURL, projectID, integrationID)

    reqBody := map[string]bool{
        "isActive": active,
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

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("status code: 404")
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to update integration status with status code: %d", resp.StatusCode)
    }

    var integration Integration
    err = json.NewDecoder(resp.Body).Decode(&integration)
    if err != nil {
        return nil, err
    }

    return &integration, nil
}

func (c *Client) GetIntegration(ctx context.Context, projectID, integrationID string) (*Integration, error) {
    url := fmt.Sprintf("%s/projects/%s/integrations/%s", c.baseURL, projectID, integrationID)

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

    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("status code: 404")
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to get integration with status code: %d", resp.StatusCode)
    }

    var integration Integration
    err = json.NewDecoder(resp.Body).Decode(&integration)
    if err != nil {
        return nil, err
    }

    return &integration, nil
}
