package knowledgebase

import (
	"context"
	"diabetes-agent-backend/config"
	"diabetes-agent-backend/dao"
	"diabetes-agent-backend/utils"
	_ "embed"
	"fmt"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const (
	baseURL            = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	embeddingModelName = "text-embedding-v4"
	collectionName     = "knowledge_doc"
	limit              = 20
)

var client *openai.LLM

//go:embed prompts/rewrite_query.txt
var rewriteQueryPrompt string

type VectorDBSearchResult struct {
	Chunk string  `json:"chunk"`
	Score float32 `json:"score"`
}

func init() {
	var err error
	client, err = openai.New(
		openai.WithEmbeddingModel(embeddingModelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(baseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create llm client: %v", err))
	}
}

func RetrieveSimilarDocuments(ctx context.Context, query, userEmail string) ([]VectorDBSearchResult, error) {
	rewrittenQuery, err := rewriteQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error rewriting query: %v", err)
	}

	embedder, err := embeddings.NewEmbedder(client)
	if err != nil {
		panic(fmt.Sprintf("Failed to create embedder: %v", err))
	}

	vector, err := embedder.EmbedQuery(ctx, rewrittenQuery)
	if err != nil {
		return nil, fmt.Errorf("error embedding query: %v", err)
	}

	searchOption := milvusclient.NewSearchOption(collectionName, limit, []entity.Vector{entity.FloatVector(vector)}).
		WithOutputFields("text").
		WithFilter("user_email == '" + userEmail + "'")

	resultSets, err := dao.MilvusClient.Search(ctx, searchOption)
	if err != nil {
		return nil, fmt.Errorf("error searching vector store: %v", err)
	}

	structedResults := make([]VectorDBSearchResult, 0)
	for _, resSet := range resultSets {
		for i := 0; i < resSet.ResultCount; i++ {
			var text string
			if textColumn := resSet.GetColumn("text"); textColumn != nil {
				if content, ok := textColumn.(*column.ColumnVarChar); ok {
					if content.Len() > 0 {
						text, _ = content.GetAsString(i)
					}
				}
			}

			structedResults = append(structedResults, VectorDBSearchResult{
				Chunk: text,
				Score: resSet.Scores[i],
			})
		}
	}

	return structedResults, nil
}

func rewriteQuery(ctx context.Context, query string) (string, error) {
	template := prompts.NewPromptTemplate(rewriteQueryPrompt, []string{"query"})
	prompt, err := template.Format(map[string]any{"query": query})
	if err != nil {
		return "", err
	}

	result, err := llms.GenerateFromSinglePrompt(ctx, client, prompt)
	if err != nil {
		return "", err
	}
	return result, nil
}
