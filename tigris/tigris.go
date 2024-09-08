package tigris

import (
    "context"
    "fmt"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func Client(ctx context.Context) (*s3.Client, error) {
    // 1. Create an aws.Config instance
    cfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to load Tigris config: %w", err)
    }

    // 2. Create an Amazon S3 client using the AWS Config instance created, "cfg"
    return s3.NewFromConfig(cfg, func(o *s3.Options){
        o.BaseEndpoint = aws.String("https://fly.storage.tigris.dev")
        o.Region = "auto"
    }), nil
}

