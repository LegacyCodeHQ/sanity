export interface RenderState {
  nextRequestID: number;
  activeRequestID: number;
  renderedDot: string | null;
}

export function createRenderState(): RenderState {
  return {
    nextRequestID: 1,
    activeRequestID: 0,
    renderedDot: null,
  };
}

export function beginRender(
  state: RenderState,
): { state: RenderState; requestID: number } {
  const requestID = state.nextRequestID;

  return {
    requestID,
    state: {
      ...state,
      nextRequestID: requestID + 1,
      activeRequestID: requestID,
    },
  };
}

export function completeRender(
  state: RenderState,
  requestID: number,
  dot: string,
): RenderState {
  if (requestID !== state.activeRequestID) {
    return state;
  }

  return {
    ...state,
    renderedDot: dot,
  };
}

export function cancelPendingRenders(state: RenderState): RenderState {
  return {
    ...state,
    activeRequestID: state.nextRequestID,
  };
}

export function clearRenderedDot(state: RenderState): RenderState {
  return {
    ...state,
    renderedDot: null,
  };
}

export function shouldRenderDot(state: RenderState, dot: string): boolean {
  return state.renderedDot !== dot;
}
