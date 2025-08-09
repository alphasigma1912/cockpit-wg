import React, { useEffect, useState } from "react";
import {
  PageSection,
  Title,
  Button,
  Alert,
  AlertGroup,
} from "@patternfly/react-core";
import { useTranslation } from "react-i18next";
import backend from "./backend";
import QRCode from "qrcode";

const Exchange: React.FC = () => {
  const { t } = useTranslation();
  const [key, setKey] = useState("");
  const [qr, setQr] = useState<string | null>(null);
  const [warn, setWarn] = useState(false);

  const loadKey = async () => {
    const k = await backend.getExchangeKey();
    setKey(k);
    const data = await QRCode.toDataURL(k);
    setQr(data);
  };

  useEffect(() => {
    loadKey();
  }, []);

  const rotate = async () => {
    await backend.rotateKeys();
    await loadKey();
    setWarn(true);
  };

  return (
    <PageSection>
      {warn && (
        <AlertGroup>
          <Alert
            variant="warning"
            title={t("exchange.redistributeWarning")}
            isInline
          />
        </AlertGroup>
      )}
      <Title headingLevel="h1">{t("exchange.title")}</Title>
      <p>
        {t("exchange.pubKeyLabel")}: {key}
      </p>
      {qr && <img src={qr} alt={t("exchange.pubKeyLabel")} />}
      <Button onClick={rotate}>{t("exchange.rotate")}</Button>
    </PageSection>
  );
};

export default Exchange;
