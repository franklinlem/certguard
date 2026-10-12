import { render, screen } from '@testing-library/react'
import { expect, test } from 'vitest'
import { App } from '../src/App'
test('informa o estágio real e não oferece monitoramento indisponível', () => {
  render(<App />)
  expect(screen.getByRole('heading', { level: 1 }).textContent).toContain('Confiança')
  expect(screen.getByRole('note').textContent).toContain('O monitoramento ainda não está ativo')
  expect(screen.queryByRole('button', { name: /cadastrar/i })).toBeNull()
})
