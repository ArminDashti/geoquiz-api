CREATE INDEX IF NOT EXISTS idx_scores_quiz_correct_created
  ON scores (quiz_type, correct DESC, created_at DESC);
