import redHatText from '@patternfly/react-core/dist/styles/assets/fonts/RedHatText/RedHatTextVF.woff2';
import redHatTextItalic from '@patternfly/react-core/dist/styles/assets/fonts/RedHatText/RedHatTextVF-Italic.woff2';
import redHatDisplay from '@patternfly/react-core/dist/styles/assets/fonts/RedHatDisplay/RedHatDisplayVF.woff2';
import redHatDisplayItalic from '@patternfly/react-core/dist/styles/assets/fonts/RedHatDisplay/RedHatDisplayVF-Italic.woff2';
import redHatMono from '@patternfly/react-core/dist/styles/assets/fonts/RedHatMono/RedHatMonoVF.woff2';
import redHatMonoItalic from '@patternfly/react-core/dist/styles/assets/fonts/RedHatMono/RedHatMonoVF-Italic.woff2';
import pfIcon from '@patternfly/react-core/dist/styles/assets/pficon/pf-v5-pficon.woff2';

const fonts = [
  redHatText,
  redHatTextItalic,
  redHatDisplay,
  redHatDisplayItalic,
  redHatMono,
  redHatMonoItalic,
  pfIcon
];

for (const href of fonts) {
  const link = document.createElement('link');
  link.rel = 'preload';
  link.as = 'font';
  link.href = href;
  link.type = 'font/woff2';
  link.crossOrigin = '';
  document.head.appendChild(link);
}
