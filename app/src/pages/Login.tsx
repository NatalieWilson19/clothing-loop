import {
  IonHeader,
  IonToolbar,
  IonTitle,
  IonContent,
  IonButton,
  IonModal,
  IonItem,
  IonLabel,
  IonInput,
  IonIcon,
  IonText,
  useIonToast,
} from "@ionic/react";
import {
  arrowForwardOutline,
  mailUnreadOutline,
  sendOutline,
} from "ionicons/icons";
import { Keyboard } from "@capacitor/keyboard";
import {
  Fragment,
  KeyboardEventHandler,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { useTranslation } from "react-i18next";
import { useHistory } from "react-router";
import toastError from "../../toastError";
import { loginEmail } from "../api";
import { StoreContext } from "../Store";

enum State {
  idle,
  error,
  success,
}

const KEYCODE_ENTER = 13;

const BETA_TESTERS = (import.meta.env.VITE_APP_BETA_TESTERS || "").split(",");
const IS_DEVELOPMENT = import.meta.env.DEV;

export default function Login(props: { isLoggedIn: boolean }) {
  const { t } = useTranslation();
  const { login } = useContext(StoreContext);
  const history = useHistory();
  const [present] = useIonToast();

  const modal = useRef<HTMLIonModalElement>(null);
  const inputEmail = useRef<HTMLIonInputElement>(null);
  const inputToken = useRef<HTMLIonInputElement>(null);

  const [showToken, setShowToken] = useState(false);
  const [sentState, setSentState] = useState(State.idle);
  const [verifyState, setVerifyState] = useState(State.idle);
  const [sentTimeout, setSentTimeout] = useState<number>();

  async function handleSendEmail() {
    if (sentState === State.success) return;
    clearTimeout(sentTimeout);
    const email = inputEmail.current?.value + "";
    if (!email) return;

    if (!IS_DEVELOPMENT && !BETA_TESTERS.includes(email)) {
      setSentState(State.error);
      toastError(
        present,
        "This app is currently being beta tested, only a select few can access it at this time"
      );
      return;
    }

    try {
      const res = await loginEmail(email + "");
      setShowToken(true);
      setSentState(State.success);
      setSentTimeout(
        setTimeout(() => setSentState(State.idle), 1000 * 60 /* 1 min */) as any
      );
      Keyboard.hide();
    } catch (err) {
      setSentState(State.error);
      toastError(present, err);
      console.error(err);
    }
  }

  function handleInputEmailEnter(e: any) {
    if (e?.keyCode === KEYCODE_ENTER) {
      handleSendEmail();
    }
  }

  function handleInputTokenEnter(e: any) {
    if (e?.keyCode === KEYCODE_ENTER) {
      handleVerifyToken();
    }
  }

  async function handleVerifyToken() {
    const token = inputToken.current?.value || "";
    if (!token) return;

    try {
      await login(token + "");
      setVerifyState(State.success);
      modal.current?.dismiss("success");
      history.replace("/settings", "select-loop");
    } catch (e: any) {
      console.error(e);
      setVerifyState(State.error);
    }
  }

  return (
    <IonModal
      ref={modal}
      isOpen={!props.isLoggedIn}
      canDismiss={async (d) => d === "success"}
    >
      <IonHeader>
        <IonToolbar>
          <IonTitle>{t("login")}</IonTitle>
        </IonToolbar>
      </IonHeader>
      <IonContent className="ion-padding">
        <IonItem lines="none">
          <IonText>{t("pleaseEnterYourEmailAddress")}</IonText>
        </IonItem>
        <IonItem lines="none">
          <IonInput
            label={t("email")!}
            labelPlacement="fixed"
            ref={inputEmail}
            type="email"
            autocomplete="on"
            autoSave="on"
            autofocus
            enterkeyhint="send"
            onKeyUp={handleInputEmailEnter}
            aria-autocomplete="list"
            required
            placeholder={t("yourEmailAddress")!}
          />
        </IonItem>
        <IonItem lines="none">
          <IonButton
            size="default"
            slot="end"
            expand="block"
            color={
              sentState === State.error
                ? "danger"
                : sentState === State.success
                ? "success"
                : "primary"
            }
            disabled={sentState === State.success}
            onClick={handleSendEmail}
          >
            Send
            {sentState === State.success ? (
              <IonIcon slot="end" icon={mailUnreadOutline} />
            ) : (
              <IonIcon slot="end" icon={sendOutline} />
            )}
          </IonButton>
        </IonItem>
        {showToken ? (
          <Fragment key="token">
            <IonItem lines="none">
              <IonText>{t("enterThePasscodeYouReceivedInYourEmail")}</IonText>
            </IonItem>
            <IonItem lines="none">
              <IonInput
                type="number"
                ref={inputToken}
                autoCorrect="off"
                placeholder="••••••"
                label={t("passcode")!}
                enterkeyhint="enter"
                onKeyUp={handleInputTokenEnter}
                labelPlacement="fixed"
              />
            </IonItem>
            <IonItem lines="none">
              <IonButton
                color={
                  verifyState === State.error
                    ? "danger"
                    : verifyState === State.success
                    ? "success"
                    : "primary"
                }
                size="default"
                disabled={verifyState === State.success}
                slot="end"
                expand="block"
                onClick={handleVerifyToken}
              >
                <IonLabel>{t("login")}</IonLabel>
                <IonIcon slot="end" icon={arrowForwardOutline} />
              </IonButton>
            </IonItem>
          </Fragment>
        ) : null}
      </IonContent>
    </IonModal>
  );
}
