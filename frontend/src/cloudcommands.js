const cloudCommands = {
    gcp: [
        { command: "gcloud compute instances list", args: ["--format", "--filter", "--limit", "--sort-by"] },
        { command: "gcloud compute instances create", args: ["--machine-type", "--zone", "--image-family", "--image-project"] },
        { command: "gcloud compute ssh", args: ["--zone", "--project", "--command"] },
        { command: "gcloud storage ls", args: ["--recursive", "--long"] },
        { command: "gcloud storage cp", args: ["--recursive"] },
        { command: "gcloud container clusters create", args: ["--num-nodes", "--zone", "--machine-type"] },
        { command: "gcloud container clusters get-credentials", args: ["--zone", "--project"] },
        { command: "gcloud functions deploy", args: ["--runtime", "--trigger-http", "--allow-unauthenticated"] },
        { command: "gcloud pubsub topics create", args: [] },
        { command: "gcloud pubsub subscriptions create", args: ["--topic"] },
    ],
    aws: [
        { command: "aws ec2 describe-instances", args: ["--filters", "--query", "--output"] },
        { command: "aws ec2 run-instances", args: ["--image-id", "--instance-type", "--key-name", "--security-group-ids"] },
        { command: "aws s3 ls", args: ["--recursive", "--human-readable", "--summarize"] },
        { command: "aws s3 cp", args: ["--recursive", "--acl"] },
        { command: "aws eks create-cluster", args: ["--name", "--role-arn", "--resources-vpc-config"] },
        { command: "aws eks get-token", args: ["--cluster-name"] },
        { command: "aws lambda create-function", args: ["--function-name", "--runtime", "--role", "--handler"] },
        { command: "aws sns create-topic", args: ["--name"] },
        { command: "aws sqs create-queue", args: ["--queue-name"] },
        { command: "aws dynamodb create-table", args: ["--table-name", "--attribute-definitions", "--key-schema", "--provisioned-throughput"] },
    ]
};

export default cloudCommands;