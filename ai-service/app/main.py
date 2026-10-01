import base64
import binascii
import os
from contextlib import asynccontextmanager

import cv2
import numpy as np
from fastapi import FastAPI, HTTPException
from insightface.app import FaceAnalysis
from pydantic import BaseModel, Field


class EnrollRequest(BaseModel):
    images: list[str] = Field(min_length=2, max_length=3)


class VerifyRequest(BaseModel):
    image: str
    enrolled_embedding: list[float] = Field(min_length=512, max_length=512)
    challenge: str
    threshold: float = Field(default=0.45, ge=0.0, le=1.0)


class FaceEngine:
    def __init__(self) -> None:
        self.detector = FaceAnalysis(
            name=os.getenv("INSIGHTFACE_MODEL", "buffalo_l"),
            providers=[os.getenv("ONNX_PROVIDER", "CPUExecutionProvider")],
        )
        self.detector.prepare(ctx_id=-1, det_size=(640, 640))

    @staticmethod
    def decode_image(encoded: str) -> np.ndarray:
        if encoded.startswith("data:"):
            encoded = encoded.split(",", 1)[1]
        try:
            raw = base64.b64decode(encoded, validate=True)
        except (ValueError, binascii.Error) as exc:
            raise HTTPException(400, "image must be valid base64") from exc
        if not raw or len(raw) > 1_048_576:
            raise HTTPException(400, "image must be between 1 byte and 1 MB")
        image = cv2.imdecode(np.frombuffer(raw, np.uint8), cv2.IMREAD_COLOR)
        if image is None:
            raise HTTPException(400, "image must be JPEG or PNG")
        return image

    def extract(self, encoded: str):
        faces = self.detector.get(self.decode_image(encoded))
        if len(faces) != 1:
            raise HTTPException(422, "exactly one face must be visible")
        face = faces[0]
        if face.det_score < 0.75:
            raise HTTPException(422, "face image quality is too low")
        embedding = np.asarray(face.normed_embedding, dtype=np.float32)
        return face, embedding / np.linalg.norm(embedding)

    @staticmethod
    def passes_pose(face, challenge: str) -> bool:
        if challenge not in {"turn_left", "turn_right"}:
            return False
        left_eye, right_eye, nose = face.kps[:3]
        eye_distance = max(float(np.linalg.norm(right_eye - left_eye)), 1.0)
        nose_offset = float((nose[0] - ((left_eye[0] + right_eye[0]) / 2)) / eye_distance)
        required_offset = float(os.getenv("LIVENESS_YAW_THRESHOLD", "0.10"))
        return nose_offset <= -required_offset if challenge == "turn_left" else nose_offset >= required_offset


engine: FaceEngine | None = None


@asynccontextmanager
async def lifespan(_: FastAPI):
    global engine
    engine = FaceEngine()
    yield
    engine = None


app = FastAPI(title="PresensiGo Face Service", version="1.0.0", lifespan=lifespan)


@app.get("/health")
def health():
    return {"status": "ok", "model_ready": engine is not None}


@app.post("/v1/faces/enroll")
def enroll(request: EnrollRequest):
    assert engine is not None
    embeddings = [engine.extract(image)[1] for image in request.images]
    centroid = np.mean(embeddings, axis=0)
    centroid /= np.linalg.norm(centroid)
    consistency = min(float(np.dot(item, centroid)) for item in embeddings)
    if consistency < 0.65:
        raise HTTPException(422, "enrollment selfies do not appear to show the same person")
    return {"embedding": centroid.tolist(), "samples": len(embeddings), "consistency": consistency}


@app.post("/v1/faces/verify")
def verify(request: VerifyRequest):
    assert engine is not None
    face, candidate = engine.extract(request.image)
    enrolled = np.asarray(request.enrolled_embedding, dtype=np.float32)
    enrolled /= max(float(np.linalg.norm(enrolled)), 1e-8)
    similarity = float(np.dot(candidate, enrolled))
    live = engine.passes_pose(face, request.challenge)
    return {
        "verified": similarity >= request.threshold and live,
        "similarity": similarity,
        "threshold": request.threshold,
        "liveness_passed": live,
        "challenge": request.challenge,
    }
