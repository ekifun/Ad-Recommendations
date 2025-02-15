Summary of Category and Embedding-Based Hybrid Ranking in the Current Ad-Recommendation System
1️⃣ Overview of the Hybrid Ranking Approach
The Ad-Recommendation System combines:
Category-Based Matching → Maps user playback history to relevant ad categories.
BERT-Based Semantic Similarity → Computes similarity between user history and ad descriptions.
Hybrid Scoring Algorithm → Balances category scores and embedding-based similarity to rank ads.
This ensures that both explicit user preferences (category mapping) and semantic relevance (BERT embeddings) influence ranking.
2️⃣ Category-Based Ranking
How It Works:
User’s playback history is retrieved from DynamoDB.
Each movie category is mapped to corresponding ad categories using the CategoryMappingTable.
Weights are assigned based on category importance.
Key Issues Found:
Some categories (e.g., Gadgets, Software) were over-represented, leading to bias.
Weight accumulation was excessive, making category scores dominate rankings.
Normalization was ineffective, allowing large category scores to overshadow embedding scores.
Improvements:
✅ Applied frequency-based normalization to prevent over-represented categories.
✅ Ensured category weights do not exceed reasonable bounds.
✅ Logged intermediate weight calculations for debugging.
3️⃣ Embedding-Based Ranking (BERT Similarity)
How It Works:
User and ad descriptions are converted into embeddings using a BERT model.
Cosine similarity is computed between user embeddings and ad embeddings.
Higher similarity means better semantic relevance of an ad to the user’s interests.
Key Issues Found:
BERT similarity had minimal effect due to the dominance of category scores.
Some embeddings were too close, leading to less variation in ranking.
Lack of diverse ads limited the impact of embeddings.
Improvements:
✅ Increased the weight of BERT similarity in the final score.
✅ Enhanced diversity in ad retrieval to improve embedding-based ranking impact.
✅ Logged embedding scores to debug their role in final ranking.
4️⃣ Hybrid Ranking Algorithm
How It Works:
Combines category-based scores and embedding-based scores.
Adjusts the weight balance to ensure both signals influence ranking.
Previous Formula (Category-Dominant):
finalScore = (0.7 * categoryScore) + (0.3 * bertScore)
New Formula (More Balanced):
finalScore = (0.4 * categoryScore) + (0.6 * bertScore)
Key Issues Found:
Category scores were too dominant, causing BERT similarity to be ignored.
Lack of ad diversity skewed ranking, favoring a few categories repeatedly.
Some categories had inflated scores, leading to bias.
Improvements:
✅ Balanced weights to let BERT embeddings contribute more to ranking.
✅ Improved normalization to reduce category dominance.
✅ Introduced diversity constraints in ad selection.
✅ Logged final scores for debugging ranking consistency.
5️⃣ Expected Impact of Changes
More diverse and relevant recommendations.
BERT embeddings now meaningfully contribute to ranking.
Prevents over-representation of a few categories.
Ad ranking changes dynamically based on user history and ad content.
6️⃣ Next Steps
Monitor logs to ensure category and embedding scores are balanced.
Test ranking with different user inputs to verify diversity improvements.
Fine-tune weight parameters (e.g., 0.4 category, 0.6 BERT) based on real data feedback.
Introduce a fallback strategy if embeddings fail to generate.
Conclusion
This update optimizes the hybrid ranking system by ensuring that both category-based and embedding-based relevance influence recommendations, reducing bias and improving diversity. 