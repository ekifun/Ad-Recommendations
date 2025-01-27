Summary of Ad Recommendation Systems and Workflow:
1. Overview
    We have implemented a system using a Playback Simulator to interact with a DynamoDB database for saving user interaction data. This data is then utilized to build an AI-based Ad Recommendation System using an embedded AI model for relevance scoring.
Key Features and Changes
2. DynamoDB Setup
2.1 Tables Created:
2.1.1 Users:
        Partition Key: user_id
        Stores user profiles and interaction history.
2.1.2 PlaybackTable:
        Partition Key: user_id
        Stores movie playback history for each user.
2.1.3 AdClickTable:
        Partition Key: user_id, Sort Key: ad_id
        Stores ad click events.
2.2 AWS Configuration:
        IAM roles were set up with dynamodb:* permissions.
        Environment variables were configured for AWS credentials and region.
3. Playback Simulator
    3.1 Simulates user playback events by submitting user_id and category to the /playback endpoint.
    3.2 Example cURL command:
        curl -X POST -H "Content-Type: application/json" -d '{
            "user_id": "12345",
            "category": "Fitness"
        }' http://localhost:8082/playback
    3.3 Writes playback history to PlaybackTable in DynamoDB.
4. Ad Click Tracking
    4.1 Tracks user clicks on ads using the /ad-click endpoint.
    4.2 Example cURL command:
        curl -X POST -H "Content-Type: application/json" -d '{
            "user_id": "12345",
            "ad_id": "1"
        }' http://localhost:8082/ad-click
    4.3 Writes ad click events to AdClickTable in DynamoDB.
5. User Management
    5.1 Enables adding, updating, deleting, and fetching user profiles using endpoints:
        /add-user
        /update-user
        /delete-user
        /get-user
    5.2 Example for adding a user:
        curl -X POST -H "Content-Type: application/json" -d '{
            "userID": "12345",
            "history": ["Tech", "Fitness"]
        }' http://localhost:8082/add-user
    5.3 Stores user data in the Users table.
6. AI-Based Ad Recommendation System
    6.1 Embedding AI Model:
        Embeddings for ads and user interactions are generated using an AI service.
        User vectors are calculated based on playback and ad-click history.
    6.2 Recommendation Process:
        Playback and ad-click history are fetched from DynamoDB.
        A user vector is computed by combining interaction embeddings.
        Ads are scored based on Cosine Similarity with the user vector.
        The top 5 ads are returned as recommendations.
    6.3 Example Recommendation Request:
        curl -X POST -H "Content-Type: application/json" -d '{
            "user_id": "12345",
            "ads": [                      
                {"ad_id": "1", "description": "Smartwatch for fitness tracking"},
                {"ad_id": "2", "description": "Healthy meal subscription service"}
            ]
        }' http://localhost:8082/recommend
    Output:
        [
            {"ad_id": "1", "category": "Tech", "description": "Latest gadgets and devices"},
            {"ad_id": "2", "category": "Fitness", "description": "Workout equipment and accessories"}
        ]
