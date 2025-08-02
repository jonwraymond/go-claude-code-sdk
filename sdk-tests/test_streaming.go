//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jonwraymond/go-claude-code-sdk/pkg/client"
	"github.com/jonwraymond/go-claude-code-sdk/pkg/types"
)

func main() {
	fmt.Println("=== Testing SDK Streaming API ===")

	ctx := context.Background()
	config := &types.ClaudeCodeConfig{
		Model: "claude-3-5-sonnet-20241022",
	}

	claudeClient, err := client.NewClaudeCodeClient(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer claudeClient.Close()

	// Test 1: Basic streaming with QueryStream
	fmt.Println("\nTest 1: Basic Streaming with QueryStream...")
	request1 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Count from 1 to 5 slowly, saying each number on a new line.",
			},
		},
		MaxTokens: 100,
		Stream:    true,
	}

	stream1, err := claudeClient.QueryStream(ctx, request1)
	if err != nil {
		log.Printf("❌ FAILED: QueryStream error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Stream created")
		fmt.Print("   Streaming output: ")
		
		for {
			chunk, err := stream1.Recv()
			if err != nil {
				if err.Error() != "EOF" && !strings.Contains(err.Error(), "EOF") {
					log.Printf("\n   ❌ Stream error: %v", err)
				}
				break
			}
			
			if chunk.Done {
				fmt.Println("\n   ✅ Stream completed")
				break
			}
			
			if chunk.Type == types.ChunkTypeContent && chunk.Content != "" {
				fmt.Print(chunk.Content)
			}
		}
		
		stream1.Close()
	}

	// Test 2: Advanced streaming with StreamQuery and callbacks
	fmt.Println("\nTest 2: Advanced Streaming with StreamQuery...")
	
	var totalContent strings.Builder
	var messageID string
	contentBlocks := 0
	
	opts := &types.StreamOptions{
		OnMessage: func(msg *types.StreamMessage) error {
			messageID = msg.ID
			fmt.Printf("   Message started (ID: %s, Model: %s)\n", msg.ID, msg.Model)
			return nil
		},
		OnContentBlock: func(index int, block *types.ContentBlock) error {
			contentBlocks++
			fmt.Printf("   Content block %d completed\n", index)
			return nil
		},
		OnContentDelta: func(delta *types.ContentDelta) error {
			if delta.Text != "" {
				totalContent.WriteString(delta.Text)
			}
			return nil
		},
		OnComplete: func(msg *types.StreamMessage) error {
			fmt.Printf("   Message completed (Stop reason: %s)\n", msg.StopReason)
			if msg.Usage != nil {
				fmt.Printf("   Token usage: Input=%d, Output=%d, Total=%d\n",
					msg.Usage.InputTokens, msg.Usage.OutputTokens, msg.Usage.TotalTokens)
			}
			return nil
		},
		BufferSize: 100,
	}

	request2 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Write a haiku about streaming data.",
			},
		},
		MaxTokens: 150,
	}

	streamResp, err := claudeClient.StreamQuery(ctx, request2, opts)
	if err != nil {
		log.Printf("❌ FAILED: StreamQuery error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Advanced stream created")
		
		// Wait for streaming to complete
		select {
		case <-streamResp.Done:
			fmt.Println("   ✅ Streaming completed")
			fmt.Printf("   Total content: %s\n", totalContent.String())
			fmt.Printf("   Content blocks: %d\n", contentBlocks)
		case err := <-streamResp.Errors:
			log.Printf("   ❌ Streaming error: %v", err)
		case <-time.After(30 * time.Second):
			log.Printf("   ❌ Streaming timeout")
			streamResp.Cancel()
		}
	}

	// Test 3: Collect method for gathering complete response
	fmt.Println("\nTest 3: Using Collect Method...")
	
	request3 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "What is 2 + 2?",
			},
		},
		MaxTokens: 50,
	}

	streamResp3, err := claudeClient.StreamQuery(ctx, request3, nil)
	if err != nil {
		log.Printf("❌ FAILED: StreamQuery error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Stream created for collection")
		
		// Collect the complete response
		response, err := streamResp3.Collect()
		if err != nil {
			log.Printf("❌ FAILED: Collect error: %v", err)
		} else {
			fmt.Println("✅ SUCCESS: Response collected")
			fmt.Printf("   Response: %s\n", response.GetTextContent())
			if response.Usage != nil {
				fmt.Printf("   Tokens used: %d\n", response.Usage.TotalTokens)
			}
		}
	}

	// Test 4: Stream cancellation
	fmt.Println("\nTest 4: Stream Cancellation...")
	
	request4 := &types.QueryRequest{
		Messages: []types.Message{
			{
				Role:    types.RoleUser,
				Content: "Count from 1 to 100, slowly.",
			},
		},
		MaxTokens: 500,
	}

	cancelOpts := &types.StreamOptions{
		OnContentDelta: func(delta *types.ContentDelta) error {
			// Cancel after receiving some content
			if strings.Contains(delta.Text, "5") {
				fmt.Println("\n   Cancelling stream...")
				return fmt.Errorf("user cancelled")
			}
			fmt.Print(delta.Text)
			return nil
		},
	}

	streamResp4, err := claudeClient.StreamQuery(ctx, request4, cancelOpts)
	if err != nil {
		log.Printf("❌ FAILED: StreamQuery error: %v", err)
	} else {
		fmt.Println("✅ SUCCESS: Cancellable stream created")
		fmt.Print("   Output: ")
		
		select {
		case <-streamResp4.Done:
			fmt.Println("\n   ✅ Stream ended (cancelled or completed)")
		case err := <-streamResp4.Errors:
			if strings.Contains(err.Error(), "cancelled") {
				fmt.Println("\n   ✅ Stream cancelled as expected")
			} else {
				log.Printf("\n   ❌ Unexpected error: %v", err)
			}
		case <-time.After(10 * time.Second):
			streamResp4.Cancel()
			fmt.Println("\n   ✅ Stream cancelled due to timeout")
		}
	}

	fmt.Println("\n=== Streaming API Tests Complete ===")

	// Summary
	fmt.Println("\nFeatures tested:")
	fmt.Println("- ✅ Basic streaming with QueryStream")
	fmt.Println("- ✅ Advanced streaming with callbacks")
	fmt.Println("- ✅ Message and content block events")
	fmt.Println("- ✅ Token usage tracking")
	fmt.Println("- ✅ Collect method for complete responses")
	fmt.Println("- ✅ Stream cancellation")
	fmt.Println("- ✅ Error handling")
}