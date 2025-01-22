from flask import Flask, request, jsonify
from sklearn.feature_extraction.text import TfidfVectorizer
from transformers import AutoTokenizer, AutoModel
import numpy as np

app = Flask(__name__)

# Initialize TF-IDF Vectorizer
vectorizer = TfidfVectorizer(stop_words='english', max_features=5000)

# Initialize BERT Model
tokenizer = AutoTokenizer.from_pretrained("bert-base-uncased")
model = AutoModel.from_pretrained("bert-base-uncased")

@app.route('/generate-tfidf', methods=['POST'])
def generate_tfidf():
    data = request.json['texts']
    tfidf_matrix = vectorizer.fit_transform(data)
    return jsonify(tfidf_matrix.toarray().tolist())

@app.route('/generate-bert', methods=['POST'])
def generate_bert():
    data = request.json['texts']
    embeddings = []
    for text in data:
        inputs = tokenizer(text, return_tensors="pt", truncation=True, padding=True, max_length=128)
        outputs = model(**inputs)
        embeddings.append(outputs.last_hidden_state.mean(dim=1).detach().numpy().tolist()[0])
    return jsonify(embeddings)

if __name__ == '__main__':
    app.run(port=5001)
