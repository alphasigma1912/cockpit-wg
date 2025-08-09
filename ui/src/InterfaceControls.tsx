import React, { useEffect, useState, useRef } from "react";
import { useTranslation } from "react-i18next";
import {
  Button,
  FormGroup,
  Modal,
  PageSection,
  Title,
} from "@patternfly/react-core";
import backend from "./backend";
import MetricsGraph from "./MetricsGraph";
import ErrorAlert from './ErrorAlert';
import { BackendError } from './errorCodes';

const InterfaceControls: React.FC = () => {
  const { t } = useTranslation();
  const [interfaces, setInterfaces] = useState<string[]>([]);
  const [selected, setSelected] = useState<string>("");
  const [status, setStatus] = useState("");
  const [lastChange, setLastChange] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState<BackendError | null>(null);
  const [confirmDown, setConfirmDown] = useState(false);
  const [metrics, setMetrics] = useState<{
    timestamps: number[];
    rx: number[];
    tx: number[];
  }>({
    timestamps: [],
    rx: [],
    tx: [],
  });
  const refreshTimer = useRef<number>();

  useEffect(() => {
    backend
      .listInterfaces()
      .then((res) => {
        const list: string[] = res.interfaces || [];
        setInterfaces(list);
        if (list.length > 0) {
          setSelected(list[0]);
        }
      })
      .catch((e: BackendError) => setError(e));
  }, []);

  const refreshStatus = () => {
    if (!selected) return;
    backend
      .getInterfaceStatus(selected)
      .then((res) => {
        setStatus(res.status);
        setLastChange(res.last_change);
        setMessage(res.message || "");
      })
      .catch((e: BackendError) => setError(e));
  };

  useEffect(() => {
    refreshStatus();
    let cancelled = false;
    const fetchMetrics = () => {
      if (!selected) return;
      backend
        .getMetrics(selected)
        .then((res) => {
          if (!cancelled) setMetrics(res);
        })
        .catch((e: BackendError) => setError(e));
    };
    fetchMetrics();
    const id = setInterval(fetchMetrics, 2000);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [selected]);

  const scheduleRefresh = () => {
    if (refreshTimer.current) {
      clearTimeout(refreshTimer.current);
    }
    refreshTimer.current = window.setTimeout(() => {
      refreshStatus();
      const id = setInterval(refreshStatus, 2000);
      setTimeout(() => clearInterval(id), 10000);
    }, 500);
  };

  const doAction = (action: "up" | "down" | "restart") => () => {
    setError(null);
    let p: Promise<any>;
    if (action === "up") p = backend.upInterface(selected);
    else if (action === "down") p = backend.downInterface(selected);
    else p = backend.restartInterface(selected);
    p.catch((e: BackendError) => setError(e)).finally(scheduleRefresh);
  };

  const confirmDownAction = () => {
    setConfirmDown(false);
    doAction("down")();
  };

  return (
    <PageSection>
      <Title headingLevel="h1">{t("interfaces.title")}</Title>
      {error && <ErrorAlert error={error} />}
      <FormGroup
        label={t("interfaces.interfaceLabel")}
        fieldId="iface-select"
        helperText={t("interfaces.interfaceHelp")}
      >
        <select
          id="iface-select"
          value={selected}
          onChange={(e) => setSelected(e.target.value)}
          aria-label={t("interfaces.selectorAria")}
        >
          {interfaces.map((n) => (
            <option key={n} value={n}>
              {n}
            </option>
          ))}
        </select>
      </FormGroup>
      <p>{t("interfaces.status", { status })}</p>
      <p>{t("interfaces.lastChange", { lastChange })}</p>
      {message && <pre>{message}</pre>}
      <MetricsGraph
        times={metrics.timestamps}
        rx={metrics.rx}
        tx={metrics.tx}
      />
      <Button variant="primary" onClick={doAction("up")} isDisabled={!selected}>
        {t("interfaces.up")}
      </Button>{" "}
      <Button
        variant="secondary"
        onClick={() => setConfirmDown(true)}
        isDisabled={!selected}
      >
        {t("interfaces.down")}
      </Button>{" "}
      <Button
        variant="secondary"
        onClick={doAction("restart")}
        isDisabled={!selected}
      >
        {t("interfaces.restart")}
      </Button>
      <Modal
        title={t("interfaces.bringDownTitle")}
        isOpen={confirmDown}
        onClose={() => setConfirmDown(false)}
        actions={[
          <Button key="confirm" variant="danger" onClick={confirmDownAction}>
            {t("interfaces.downConfirm")}
          </Button>,
          <Button
            key="cancel"
            variant="secondary"
            onClick={() => setConfirmDown(false)}
          >
            {t("interfaces.cancel")}
          </Button>,
        ]}
      >
        {t("interfaces.modalBody")}
      </Modal>
    </PageSection>
  );
};

export default InterfaceControls;
