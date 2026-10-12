'use strict';

const express = require('express');

const app = express();
app.use(express.json({ limit: '1mb' }));

function validateMatrix(matrix, name) {
  if (!Array.isArray(matrix) || matrix.length === 0 ||
      !Array.isArray(matrix[0]) || matrix[0].length === 0) {
    throw new Error(`${name} must be a non-empty matrix`);
  }
  const cols = matrix[0].length;
  for (let i = 0; i < matrix.length; i += 1) {
    if (!Array.isArray(matrix[i]) || matrix[i].length !== cols) {
      throw new Error(`${name} must be rectangular`);
    }
    for (const value of matrix[i]) {
      if (typeof value !== 'number' || !Number.isFinite(value)) {
        throw new Error(`${name} must contain only finite numbers`);
      }
    }
  }
}

function matrixStats(matrix) {
  const values = matrix.flat();
  const sum = values.reduce((acc, value) => acc + value, 0);
  const diagonal = matrix.length === matrix[0].length &&
    matrix.every((row, i) => row.every((value, j) => i === j || value === 0));
  return {
    rows: matrix.length,
    columns: matrix[0].length,
    max: Math.max(...values),
    min: Math.min(...values),
    average: sum / values.length,
    sum,
    isDiagonal: diagonal
  };
}

function combinedStats(matrices) {
  const values = Object.values(matrices).flat(2);
  const sum = values.reduce((acc, value) => acc + value, 0);
  return {
    valueCount: values.length,
    max: Math.max(...values),
    min: Math.min(...values),
    average: sum / values.length,
    sum
  };
}

app.get('/health', (_req, res) => {
  res.json({ status: 'ok', service: 'node-api' });
});

app.post('/stats', (req, res) => {
  try {
    const matrices = req.body && req.body.matrices;
    if (!matrices || typeof matrices !== 'object' || Array.isArray(matrices) ||
        Object.keys(matrices).length === 0) {
      return res.status(400).json({ error: 'matrices must be a non-empty object of matrices' });
    }
    for (const [name, matrix] of Object.entries(matrices)) {
      validateMatrix(matrix, name);
    }
    const perMatrix = {};
    for (const [name, matrix] of Object.entries(matrices)) {
      perMatrix[name] = matrixStats(matrix);
    }
    return res.json({ perMatrix, global: combinedStats(matrices) });
  } catch (error) {
    return res.status(400).json({ error: error.message });
  }
});

const port = Number(process.env.PORT || 3000);
app.listen(port, '0.0.0.0', () => {
  console.log(`Node.js API listening on port ${port}`);
});
