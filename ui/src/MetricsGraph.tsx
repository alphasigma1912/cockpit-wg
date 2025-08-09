import React from 'react';
import { useTranslation } from 'react-i18next';

interface Props {
  times: number[];
  rx: number[];
  tx: number[];
}

const MetricsGraph: React.FC<Props> = ({ times, rx, tx }) => {
  const { t } = useTranslation();
  if (times.length === 0) {
    return <svg role="img" aria-label={t('metrics.noData')} />;
  }
  const width = 400;
  const height = 100;
  const minT = times[0];
  const maxT = times[times.length - 1] || minT + 1;
  const spanT = maxT - minT || 1;
  const maxVal = Math.max(...rx, ...tx, 1);

  const buildPath = (vals: number[]) =>
    vals
      .map((v, i) => {
        const x = ((times[i] - minT) / spanT) * width;
        const y = height - (v / maxVal) * height;
        return `${i === 0 ? 'M' : 'L'}${x},${y}`;
      })
      .join(' ');

  const rxPath = buildPath(rx);
  const txPath = buildPath(tx);

  return (
    <svg
      width="100%"
      height="100"
      viewBox={`0 0 ${width} ${height}`}
      role="img"
      aria-labelledby="traffic-title"
      aria-describedby="traffic-desc"
    >
      <title id="traffic-title">{t('metrics.title')}</title>
      <desc id="traffic-desc">{t('metrics.desc')}</desc>
      <path
        d={rxPath}
        fill="none"
        stroke="var(--pf-v5-global--palette--green-500)"
        strokeWidth="2"
      />
      <path
        d={txPath}
        fill="none"
        stroke="var(--pf-v5-global--palette--blue-500)"
        strokeWidth="2"
        strokeDasharray="4"
      />
    </svg>
  );
};

export default MetricsGraph;
