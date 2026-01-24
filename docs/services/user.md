# User

Access user information and subscription details.

## Get User Info

```go
user, err := client.User().GetInfo(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("User ID: %s\n", user.UserID)
fmt.Printf("First Name: %s\n", user.FirstName)
```

## User Object

| Field | Description |
|-------|-------------|
| `UserID` | Unique identifier |
| `FirstName` | User's first name |
| `IsNewUser` | Whether user is new |
| `CanUseDelayedPaymentMethods` | Payment options available |

## Get Subscription

```go
sub, err := client.User().GetSubscription(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Tier: %s\n", sub.Tier)
fmt.Printf("Characters Used: %d\n", sub.CharacterCount)
fmt.Printf("Character Limit: %d\n", sub.CharacterLimit)
fmt.Printf("Remaining: %d\n", sub.CharactersRemaining())
```

## Subscription Object

| Field | Description |
|-------|-------------|
| `Tier` | Subscription tier |
| `CharacterCount` | Characters used this period |
| `CharacterLimit` | Maximum characters allowed |
| `Status` | Subscription status |
| `NextBillingDate` | Next billing date |
| `Currency` | Billing currency |
| `InvoiceInfo` | Invoice details |

## Check Characters Remaining

```go
sub, _ := client.User().GetSubscription(ctx)

remaining := sub.CharactersRemaining()
if remaining < 1000 {
    fmt.Println("Warning: Low character balance!")
}
```

## Subscription Tiers

| Tier | Characters/Month |
|------|-----------------|
| Free | 10,000 |
| Starter | 30,000 |
| Creator | 100,000 |
| Pro | 500,000 |
| Scale | 2,000,000 |
| Enterprise | Custom |

## Pre-Generation Check

Always check before generating audio:

```go
func generateSafely(client *elevenlabs.Client, text string) error {
    sub, err := client.User().GetSubscription(context.Background())
    if err != nil {
        return err
    }

    charCount := len(text)
    if sub.CharactersRemaining() < charCount {
        return fmt.Errorf("insufficient characters: need %d, have %d",
            charCount, sub.CharactersRemaining())
    }

    // Safe to generate
    _, err = client.TextToSpeech().Simple(context.Background(), voiceID, text)
    return err
}
```

## Monitor Usage

```go
func printUsageReport(client *elevenlabs.Client) {
    sub, _ := client.User().GetSubscription(context.Background())

    used := sub.CharacterCount
    limit := sub.CharacterLimit
    remaining := sub.CharactersRemaining()
    pct := float64(used) / float64(limit) * 100

    fmt.Printf("Usage Report\n")
    fmt.Printf("============\n")
    fmt.Printf("Tier: %s\n", sub.Tier)
    fmt.Printf("Used: %d / %d (%.1f%%)\n", used, limit, pct)
    fmt.Printf("Remaining: %d\n", remaining)
}
```
