/*
Copyright © 2023 Natalie Somersall
*/
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
)

func main() {
	// -h flag or no arguments provided
	if len(os.Args) < 1 || os.Args[1] == "-h" {
		fmt.Println("Usage: gh org-admin-promote GITHUB_ENTERPRISE_SLUG")
		fmt.Println("Promotes the authenticated user to admin for all organizations in the specified enterprise")
		fmt.Println("GH_TOKEN requires the following scopes: admin:enterprise, admin:org")
		fmt.Println("See https://cli.github.com/manual/gh_auth_login to add scopes to gh cli!")
		os.Exit(0)
	}

	// Get the enterprise slug from args
	enterpriseSlug := os.Args[1]

	// Get the hostname from the environment variable, otherwise default to github.com
	hostname := os.Getenv("GH_HOST")
	if hostname == "" {
		hostname = "github.com"
	}

	// Create a GraphQL client using the hostname from the gh cli
	opts := api.ClientOptions{
		Host: hostname,
	}
	client, err := api.NewGraphQLClient(opts)
	if err != nil {
		log.Fatal(err)
	}

	// Get the enterprise ID from the enterprise slug
	var enterpriseIDQuery struct {
		Enterprise struct {
			ID string `graphql:"id"`
		} `graphql:"enterprise(slug: $slug)"`
	}
	variables := map[string]interface{}{
		"slug": graphql.String(enterpriseSlug),
	}
	err = client.Query("EnterpriseID", &enterpriseIDQuery, variables)
	if err != nil {
		log.Fatal(err)
	}
	enterpriseID := enterpriseIDQuery.Enterprise.ID

	// Get a total count of organizations in the enterprise
	var orgCountQuery struct {
		Enterprise struct {
			Organizations struct {
				TotalCount int `graphql:"totalCount"`
			} `graphql:"organizations"`
		} `graphql:"enterprise(slug: $slug)"`
	}
	fmt.Printf("Getting total count of organizations in %s...\n", enterpriseSlug)
	variables = map[string]interface{}{
		"slug": graphql.String(enterpriseSlug),
	}
	err = client.Query("OrgCount", &orgCountQuery, variables)
	if err != nil {
		log.Fatal(err)
	}
	orgCount := orgCountQuery.Enterprise.Organizations.TotalCount
	fmt.Printf("Total count of organizations in %s: %d\n", enterpriseSlug, orgCount)

	// Create a CSV file
	csvFile, err := os.Create("all_orgs.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Write CSV header
	err = writer.Write([]string{"ID", "CreatedAt", "Login", "Email", "ViewerCanAdminister", "ViewerIsAMember", "Repo_TotalCount", "Repo_TotalDiskUsage"})
	if err != nil {
		log.Fatal(err)
	}

	// Get a list of organizations in the enterprise
	var orgListQuery struct {
		Enterprise struct {
			Organizations struct {
				Edges []struct {
					Node struct {
						ID                  string `graphql:"id"`
						CreatedAt           string `graphql:"createdAt"`
						Login               string `graphql:"login"`
						Email               string `graphql:"email"`
						ViewerCanAdminister bool   `graphql:"viewerCanAdminister"`
						ViewerIsAMember     bool   `graphql:"viewerIsAMember"`
						ArchivedAt          string `graphql:"archivedAt"`
						Repositories        struct {
							TotalCount     int `graphql:"totalCount"`
							TotalDiskUsage int `graphql:"totalDiskUsage"`
						} `graphql:"repositories"`
					} `graphql:"node"`
				} `graphql:"edges"`
				PageInfo struct {
					EndCursor   string `graphql:"endCursor"`
					HasNextPage bool   `graphql:"hasNextPage"`
				}
			} `graphql:"organizations(first: 100, after: $cursor)"`
		} `graphql:"enterprise(slug: $slug)"`
	}
	fmt.Printf("Getting list of organizations in %s...\n", enterpriseSlug)
	variables = map[string]interface{}{
		"slug":   graphql.String(enterpriseSlug),
		"cursor": (*graphql.String)(nil),
	}
	page := 1
	for {
		if err := client.Query("OrgList", &orgListQuery, variables); err != nil {
			log.Fatal(err)
		}

		// Write each organization to the CSV file
		for _, org := range orgListQuery.Enterprise.Organizations.Edges {
			err = writer.Write([]string{org.Node.ID, org.Node.CreatedAt, org.Node.Login, org.Node.Email, fmt.Sprintf("%t", org.Node.ViewerCanAdminister), fmt.Sprintf("%t", org.Node.ViewerIsAMember), org.Node.ArchivedAt, fmt.Sprintf("%d", org.Node.Repositories.TotalCount), fmt.Sprintf("%d", org.Node.Repositories.TotalDiskUsage)})
			if err != nil {
				log.Fatal(err)
			}
		}

		// Promote this user to enterprise admin for all organizations where ViewerCanAdminister is false
		for _, org := range orgListQuery.Enterprise.Organizations.Edges {
			if !org.Node.ViewerCanAdminister {
				fmt.Printf("Promoting user to admin for %s...\n", org.Node.Login)
				var promoteAdmin struct {
					UpdateEnterpriseOwnerOrganizationRole struct {
						ClientMutationId string
					} `graphql:"updateEnterpriseOwnerOrganizationRole(input: $input)"`
				}

				type UpdateEnterpriseOwnerOrganizationRoleInput struct {
					EnterpriseId     graphql.ID     `json:"enterpriseId"`
					OrganizationId   graphql.ID     `json:"organizationId"`
					OrganizationRole graphql.String `json:"organizationRole"`
				}

				variables := map[string]interface{}{
					"input": UpdateEnterpriseOwnerOrganizationRoleInput{
						EnterpriseId:     graphql.ID(enterpriseID),
						OrganizationId:   graphql.ID(org.Node.ID),
						OrganizationRole: graphql.String("OWNER"),
					},
				}

				err = client.Mutate("PromoteAdmin", &promoteAdmin, variables)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("User promoted to admin for %s\n", org.Node.Login)
			}
		}

		// If there are no more pages, break out of the loop
		if !orgListQuery.Enterprise.Organizations.PageInfo.HasNextPage {
			break
		}

		// Otherwise, update the cursor and page number
		variables["cursor"] = graphql.String(orgListQuery.Enterprise.Organizations.PageInfo.EndCursor)
		page++
	}

	// Close the CSV file
	csvFile.Close()

}
