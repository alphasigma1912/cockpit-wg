export const CodePackageManagerFailure = 1001;
export const CodeValidationFailed = 1002;
export const CodePermissionDenied = 1003;
export const CodeMetricsUnavailable = 1004;

export interface BackendError {
  code: number;
  message: string;
  details?: string;
  timestamp?: number;
  trace?: string;
}

export const errorMessages: Record<number, string> = {
  [CodePackageManagerFailure]: 'errors.packageManagerFailed',
  [CodeValidationFailed]: 'errors.validationFailed',
  [CodePermissionDenied]: 'errors.permissionDenied',
  [CodeMetricsUnavailable]: 'errors.metricsUnavailable',
};
