package knowledgebase

import (
	"context"
	"diabetes-agent-server/config"
	"diabetes-agent-server/dao"
	"diabetes-agent-server/utils"
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const (
	llmName            = "qwen-plus"
	embeddingModelName = "text-embedding-v4"
	collectionName     = "knowledge_doc"
	scoreThreshold     = 0.5
	limit              = 20
)

var modelClient *openai.LLM

//go:embed prompts/rewrite_query.txt
var rewriteQueryPrompt string

type VectorDBSearchResult struct {
	Chunk string  `json:"chunk"`
	Score float32 `json:"score"`
}

func init() {
	var err error
	modelClient, err = openai.New(
		openai.WithModel(llmName),
		openai.WithEmbeddingModel(embeddingModelName),
		openai.WithToken(config.Cfg.Model.APIKey),
		openai.WithBaseURL(config.Cfg.Model.BaseURL),
		openai.WithHTTPClient(utils.GlobalHTTPClient),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create model client: %v", err))
	}
}

func RetrieveSimilarDocuments(ctx context.Context, query, email string) []VectorDBSearchResult {
	rewrittenQuery, err := rewriteQuery(ctx, query)
	if err != nil {
		slog.Error("error rewriting query", "err", err)
		return nil
	}

	embedder, err := embeddings.NewEmbedder(modelClient)
	if err != nil {
		slog.Error("failed to create embedder", "err", err)
		return nil
	}

	vector, err := embedder.EmbedQuery(ctx, rewrittenQuery)
	if err != nil {
		slog.Error("error embedding query", "err", err)
		return nil
	}

	searchOption := milvusclient.NewSearchOption(collectionName, limit, []entity.Vector{entity.FloatVector(vector)}).
		WithOutputFields("text").
		WithFilter("user_email == '" + email + "'")

	resultSets, err := dao.MilvusClient.Search(ctx, searchOption)
	if err != nil {
		slog.Error("error searching vector stor", "err", err)
		return nil
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
			if resSet.Scores[i] < scoreThreshold {
				continue
			}
			structedResults = append(structedResults, VectorDBSearchResult{
				Chunk: text,
				Score: resSet.Scores[i],
			})
		}
	}

	return structedResults
}

func rewriteQuery(ctx context.Context, query string) (string, error) {
	template := prompts.NewPromptTemplate(rewriteQueryPrompt, []string{"query"})
	prompt, err := template.Format(map[string]any{"query": query})
	if err != nil {
		return "", err
	}

	result, err := llms.GenerateFromSinglePrompt(ctx, modelClient, prompt)
	if err != nil {
		return "", err
	}
	return result, nil
}
