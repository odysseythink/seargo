import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { Icon } from './Icon';

describe('Icon', () => {
  it('renders an svg for known icons', () => {
    const { container } = render(<Icon name="search" />);
    const svg = container.querySelector('svg');
    expect(svg).not.toBeNull();
    expect(svg!.getAttribute('aria-hidden')).toBe('true');
  });
  it('renders with label for accessibility', () => {
    const { container } = render(<Icon name="close" label="Close" />);
    const svg = container.querySelector('svg');
    expect(svg!.getAttribute('aria-hidden')).toBeNull();
    expect(svg!.getAttribute('role')).toBe('img');
  });
  it('applies size class', () => {
    const { container } = render(<Icon name="search" size="big" />);
    const svg = container.querySelector('svg');
    expect(svg!.classList.contains('sxng-icon-big')).toBe(true);
  });
  it('returns null for unknown icon name', () => {
    const { container } = render(<Icon name={'unknown' as any} />);
    expect(container.querySelector('svg')).toBeNull();
  });
});
