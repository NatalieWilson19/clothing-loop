import {
  CreateAnimation,
  IonContent,
  IonIcon,
  IonPage,
  IonText,
} from "@ionic/react";
import { syncOutline } from "ionicons/icons";
import { useTranslation } from "react-i18next";

export default function Loading() {
  const { t } = useTranslation();
  return (
    <IonPage>
      <IonContent fullscreen>
        <div
          style={{
            display: "flex",
            height: "100%",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <CreateAnimation
            duration={1000}
            iterations={Infinity}
            play
            fromTo={{
              property: "transform",
              fromValue: "rotate(0deg)",
              toValue: "rotate(360deg)",
            }}
          >
            <IonIcon size="large" className="ion-margin" icon={syncOutline} />
          </CreateAnimation>
          <IonText className="ion-text-bold">{t("loading...")}</IonText>
        </div>
      </IonContent>
    </IonPage>
  );
}
