import { describe, it, expect } from 'vitest';
import {
  errorMessages,
  CodePackageManagerFailure,
  CodeValidationFailed,
  CodePermissionDenied,
  CodeMetricsUnavailable,
} from './errorCodes';
import './i18n';

import i18n from './i18n';

describe('errorCodes mapping', () => {
  it('provides messages', () => {
    expect(i18n.t(errorMessages[CodePackageManagerFailure])).toBe('Failed to install packages');
    expect(i18n.t(errorMessages[CodeValidationFailed])).toBe('Configuration validation failed');
    expect(i18n.t(errorMessages[CodePermissionDenied])).toBe('Permission denied');
    expect(i18n.t(errorMessages[CodeMetricsUnavailable])).toBe('Metrics provider unavailable');
  });
});
