from flask import Flask, request, jsonify
from sklearn.feature_extraction.text import TfidfVectorizer
from transformers import AutoTokenizer, AutoModel
import torch
import numpy as np

app = Flask(__name__)

# Initialize TF-IDF Vectorizer
vectorizer = TfidfVectorizer(stop_words='english', max_features=5000)

# Initialize BERT Model
tokenizer = AutoTokenizer.from_pretrained("bert-base-uncased")
model = AutoModel.from_pretrained("bert-base-uncased")
model.eval()  # Ensure model is in evaluation mode

@app.route('/generate-tfidf', methods=['POST'])
def generate_tfidf():
    data = request.json['texts']
    tfidf_matrix = vectorizer.fit_transform(data)
    return jsonify(tfidf_matrix.toarray().tolist())

@app.route('/generate-bert', methods=['POST'])
def generate_bert():
    data = request.json['texts']
    print("Received Data:", data)  # Debugging
    
    embeddings = []
    for text in data:
        inputs = tokenizer([text], padding="max_length", truncation=True, max_length=128, return_tensors="pt")
        with torch.no_grad():  # Disables gradient tracking
            outputs = model(**inputs)
        embedding = outputs.last_hidden_state.mean(dim=1).detach().cpu().numpy().tolist()[0]
        embeddings.append(embedding)
    
    return jsonify(embeddings)

if __name__ == '__main__':
    app.run(port=5001, threaded=False)  # Running single-threaded for stability
