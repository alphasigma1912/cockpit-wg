import React from 'react';

interface Props {
  times: number[];
  rx: number[];
  tx: number[];
}

const MetricsGraph: React.FC<Props> = ({ times, rx, tx }) => {
  if (times.length === 0) {
    return <svg role="img" aria-label="No traffic data" />;
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
    >
      <title id="traffic-title">Interface traffic (bytes per second)</title>
      <path
        d={rxPath}
        fill="none"
        stroke="var(--pf-v5-global--palette--green-400)"
        strokeWidth="2"
      />
      <path
        d={txPath}
        fill="none"
        stroke="var(--pf-v5-global--palette--blue-400)"
        strokeWidth="2"
      />
    </svg>
  );
};

export default MetricsGraph;
