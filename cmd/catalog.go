package cmd

import (
        "github.com/spf13/cobra"
        "github.com/olekukonko/tablewriter"
        "fmt"
        "net/http"
        "encoding/json"
        "os"
        "sort"
)

const catalog = "https://s3-eu-west-1.amazonaws.com/dev.stack.foundation/catalog/catalog.json"

type image struct {
        Name        string `json:"name"`
        Description string `json:"description"`
}

type images []image

func (s images) Len() int {
        return len(s)
}
func (s images) Swap(i, j int) {
        s[i], s[j] = s[j], s[i]
}

type byName struct{ images }

func (s byName) Less(i, j int) bool {
        return s.images[i].Name < s.images[j].Name
}

var catalogCmd = &cobra.Command{
        Use:   "catalog",
        Short: "List offical Docker images",
        Long:  `List all official Docker images to use as a base image for workflows.

Shows a list of all official Docker images that are available to use as a base image for workflows.`,
        Run: func(command *cobra.Command, args []string) {
                response, err := http.Get(catalog)
                if err != nil {
                        fmt.Printf("Error getting list of official Docker images. Check your internet connectivity and try again later.")
                        return
                }
                defer response.Body.Close()

                if response.StatusCode != 200 {
                        fmt.Printf("Error getting list of official Docker images. Check your internet connectivity and try again later.")
                        return
                }

                var images images

                err = json.NewDecoder(response.Body).Decode(&images)
                if err != nil {
                        panic(err)
                }

                sort.Sort(byName{images })

                table := tablewriter.NewWriter(os.Stdout)
                table.SetHeader([]string{"Image", "Description"})
                table.SetBorder(false)
                table.SetColumnSeparator("")
                table.SetColWidth(65)
                table.SetHeaderLine(false)

                for _, image := range images {
                        table.Append([]string{image.Name, image.Description})
                }

                table.Render()
        },
}

func init() {
        RootCmd.AddCommand(catalogCmd)
}
