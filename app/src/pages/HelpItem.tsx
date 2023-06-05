import {
  IonHeader,
  IonToolbar,
  IonContent,
  IonText,
  IonPage,
  IonBackButton,
  IonButtons,
} from "@ionic/react";
import { useContext, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { RouteComponentProps } from "react-router";
import { StoreContext } from "../Store";

export interface FaqListItem {
  Title: string;
  "Title 2": string;
  "Short explanation": string;
  "Paragraph 1": string;
  "Paragraph 2": string;
  "Paragraph 3": string;
}

export default function HelpItem({
  match,
}: RouteComponentProps<{ index: string }>) {
  const { t } = useTranslation();
  const { chain } = useContext(StoreContext);
  const data = t("list", { ns: "faq", returnObjects: true }) as any[];

  const item = useMemo(() => {
    let index = parseInt(match.params.index, 10);

    if (chain && chain.rules_override) {
      const json = JSON.parse(chain.rules_override);
      return json[index];
    }
    return data[index] as FaqListItem;
  }, [match.params.index, chain]);

  return (
    <IonPage>
      <IonHeader translucent>
        <IonToolbar>
          <IonButtons slot="start">
            <IonBackButton>{t("back")}</IonBackButton>
          </IonButtons>
        </IonToolbar>
      </IonHeader>
      <IonContent className="ion-padding">
        <IonText>
          <h1 style={{ marginTop: 0, fontSize: 30, fontWeight: "bold" }}>
            {item.Title}
          </h1>
          <h2>{item["Short explanation"]}</h2>
          <p>{item["Paragraph 1"]}</p>
          <p>{item["Paragraph 2"]}</p>
          <p>{item["Paragraph 3"]}</p>
        </IonText>
      </IonContent>
    </IonPage>
  );
}
