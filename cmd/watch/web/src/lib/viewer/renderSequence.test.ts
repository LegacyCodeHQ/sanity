import { describe, expect, it } from 'vitest';
import {
  beginRender,
  cancelPendingRenders,
  completeRender,
  type RenderState,
} from './renderSequence';

function initialState(): RenderState {
  return {
    nextRequestID: 1,
    activeRequestID: 0,
    renderedDot: null,
  };
}

describe('renderSequence', () => {
  it('ignores stale render completion from an older request', () => {
    let state = initialState();

    const first = beginRender(state);
    state = first.state;

    const second = beginRender(state);
    state = second.state;

    state = completeRender(state, second.requestID, 'digraph newest {}');
    state = completeRender(state, first.requestID, 'digraph stale {}');

    expect(state.renderedDot).toBe('digraph newest {}');
  });

  it('drops completion after pending renders are canceled', () => {
    let state = initialState();

    const inFlight = beginRender(state);
    state = inFlight.state;

    state = cancelPendingRenders(state);
    state = completeRender(state, inFlight.requestID, 'digraph old {}');

    expect(state.renderedDot).toBe(null);
  });
});
